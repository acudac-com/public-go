package cx

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/acudac-com/public-go/kms"
	"github.com/acudac-com/public-go/mux"
	"github.com/acudac-com/public-go/timex"
)

const AuthCookie string = "auth"

func Gateway(kms *kms.Kms, refreshFunc func(cx *Cx) error) mux.Gateway[*Cx] {
	return func(w http.ResponseWriter, r *http.Request, handler mux.Handler[*Cx], middleware ...*mux.Middleware[*Cx]) (err error) {
		ctx, now := timex.Now(r.Context())
		requestId := strconv.FormatInt(now.UnixNano(), 36)
		cx := &Cx{ctx, kms, &auth{}, requestId, r, false}
		cookieVal, err := r.Cookie(AuthCookie)
		if err == nil {
			if err = kms.JsonUnhashB64(cx, []byte(cookieVal.Value), cx.auth); err != nil {
				cx.Warn("unhashing auth cookie", "error", err, "requestId", requestId)
				cx.ClearAuth(w)
			} else if cx.auth.Expiry.Before(cx.StartTime()) {
				if err = refreshFunc(cx); err != nil {
					cx.Error("error refreshing auth", err)
					return mux.InternalServerErr("")
				}
			}
		}
		defer func() {
			if rec := recover(); rec != nil {
				cx.Error("panic recovered", "method", r.Method, "stack", string(debug.Stack()))
				err = mux.InternalServerErr("Internal server error")
			}
		}()
		return handler(cx, w, r)
	}
}

type Cx struct {
	context.Context
	kms        *kms.Kms
	auth       *auth
	requestId  string
	request    *http.Request
	authEdited bool
}

type auth struct {
	User    *user     `json:"U,omitempty"` // The user details if the requester is a user
	Account *account  `json:"A,omitempty"` // The account that the user/machine is accessing
	Expiry  time.Time `json:"E,omitempty"` // When this info expires and needs to be refreshed
}

type account struct {
	Id   string `json:"I,omitempty"` // The account id that the user/machine's access is for.
	Type string `json:"T,omitempty"` // The type of account (e.g. "" (default), "Starter", "Enterprise" or "Admin")
	Role string `json:"R,omitempty"` // The role of the user/machine in this account, e.g. "Owner".
}

type user struct {
	Id       string `json:"I,omitempty"` // The unique identifier for the user
	DeviceId string `json:"D,omitempty"` // The device id of the user
	Machine  bool   `json:"M,omitempty"` // Whether the user is a machine
	Key      string `json:"K,omitempty"` // The key to refresh access for this user on this device
}

func (c *Cx) Warn(msg string, args ...any) {
	slog.Warn(msg, c.slogArgs(args)...)
}

func (c *Cx) Error(msg string, args ...any) {
	slog.Error(msg, c.slogArgs(args)...)
}

func (c *Cx) Info(msg string, args ...any) {
	slog.Error(msg, c.slogArgs(args)...)
}

func (c *Cx) slogArgs(args ...any) []any {
	args = append(args, "requestId", c.requestId)
	if c.auth.User != nil {
		args = append(args, slog.Any("auth", c.auth))
	} else {
		args = append(args, "anonymous", true)
	}
	return args
}

func (c *Cx) StartTime() time.Time {
	_, now := timex.Now(c.Context)
	return now
}

func (c *Cx) User() *user {
	return c.auth.User
}

func (c *Cx) Account() *account {
	return c.auth.Account
}

func (c *Cx) SetUser(w http.ResponseWriter, user *user, account *account, expireIn time.Duration) {
	if c.authEdited {
		panic("cx.Cx.SetUser() called after auth was modified")
	}
	c.auth.User = user
	c.SetAccount(w, account, expireIn)
}

// Do not call this outside the "/auth" path
func (c *Cx) SetAccount(w http.ResponseWriter, account *account, expireIn time.Duration) {
	if c.authEdited {
		panic("cx.Cx.SetAccount() called after auth was modified")
	}
	c.auth.Account = account
	c.auth.Expiry = c.StartTime().Add(expireIn)
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookie,
		Value:    string(c.kms.B64HashJson(c, c.auth)),
		Path:     "/",
		MaxAge:   400 * 24 * 60 * 60,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// Do not call this outside the "/auth" path
func (c *Cx) ClearAuth(w http.ResponseWriter) {
	if c.authEdited {
		panic("cx.Cx.ClearAuth() called after auth was modified")
	}
	c.auth = &auth{}
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete the cookie
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
