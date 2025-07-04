package mux_test

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"testing"
	"time"

	"github.com/acudac-com/public-go/mux"
	"github.com/acudac-com/public-go/timex"
)

type Account struct {
	Id   string `json:"I,omitempty"` // The account id that the user/machine's access is for.
	Type string `json:"T,omitempty"` // The type of account (e.g. "" (default), "Starter", "Enterprise" or "Admin")
	Role string `json:"R,omitempty"` // The role of the user/machine in this account, e.g. "Owner".
}

type User struct {
	Id           string `json:"I,omitempty"` // The unique identifier for the user
	DeviceId     string `json:"D,omitempty"` // The device id of the user
	RefreshToken string `json:"R,omitempty"` // The token to refresh access for this device
}

type Machine struct {
	Id  string `json:"I,omitempty"` // The unique identifier for the machine
	Key string `json:"K,omitempty"` // The API key for the machine to refresh access
}

type requester struct {
	User    *User     `json:"U,omitempty"` // The user details if the requester is a user
	Machine *Machine  `json:"M,omitempty"` // The machine details if the requester is a machine
	Account *Account  `json:"A,omitempty"` // The account that the user/machine is accessing
	Expiry  time.Time `json:"E,omitempty"` // When this info expires and needs to be refreshed
	updated bool
}

type Cx struct {
	context.Context
	requester  *requester
	Authorized bool
}

func (c *Cx) StartTime() time.Time {
	newCtx, now := timex.Now(c.Context)
	c.Context = newCtx
	return now
}

func (c *Cx) User() *User {
	return c.requester.User
}

func (c *Cx) Machine() *Machine {
	return c.requester.Machine
}

func (c *Cx) Account() *Account {
	return c.requester.Account
}

// Returns whether the requester is a signed in user or machine
func (c *Cx) Authenticated() bool {
	return c.requester.User != nil || c.requester.Machine != nil
}

// Returns whether the access info (i.e. the User/Machine and Account) has
// expired and needs to be refreshed. Always returns false if the requester is
// not authenticated.
func (c *Cx) SetUser(user *User, account *Account, expireIn time.Duration) {
	if user == nil {
		panic("mux.Cx.SetUser() called with nil user")
	}
	if account == nil {
		panic("mux.Cx.SetUser() called with nil account")
	}
	c.requester.User = user
	c.SetAccount(account, expireIn)
}

func (c *Cx) SetAccount(account *Account, expireIn time.Duration) {
	if account == nil {
		panic("mux.Cx.SetAccount() called with nil account")
	}
	c.requester.Account = account
	c.requester.Expiry = c.StartTime().Add(expireIn)
	c.requester.updated = true
}

func (c *Cx) ClearRequester() {
	c.requester = nil
	c.requester.updated = true
}

func TestHandle(t *testing.T) {
	m := mux.New(gateway)
	m.Get("/hello", hello, authMiddleware)
	m.Get("/missing", missing, unauthMiddleware, authMiddleware)
	m.Get("/nilptr", nilptr, authMiddleware)
	m.ListenAndServe(":8080")
}

func gateway(w http.ResponseWriter, r *http.Request, handler mux.Handler[*Cx], middleware ...mux.Handler[*Cx]) (err error) {
	cx := &Cx{requester: &requester{User: &User{Id: "12345"}}}
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("panic", "method", r.Method, "user", cx.User(), "stack", string(debug.Stack()))
			err = mux.InternalServerErr("Internal server error")
		}
	}()
	for _, mw := range middleware {
		if err = mw(cx, w, r); err != nil {
			return err
		}
	}
	if !cx.Authorized {
		if r.Method == http.MethodGet {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		} else {
			return mux.ForbiddenErr(fmt.Sprintf("missing necessary permissions for %s %s", r.Method, r.URL.Path))
		}
		return
	}
	err = handler(cx, w, r)
	slog.Info("", "method", r.Method)
	return err
}

func authMiddleware(cx *Cx, w http.ResponseWriter, r *http.Request) error {
	cx.Authorized = true
	return nil
}

func unauthMiddleware(cx *Cx, w http.ResponseWriter, r *http.Request) error {
	cx.Authorized = false
	return nil
}

func hello(cx *Cx, w http.ResponseWriter, r *http.Request) error {
	w.Write([]byte("hello"))
	return nil
}

func missing(cx *Cx, w http.ResponseWriter, r *http.Request) error {
	return mux.NotFoundErr("not here")
}

var t *time.Time

func nilptr(cx *Cx, w http.ResponseWriter, r *http.Request) error {
	w.Write([]byte(t.Format("2006-01-02 15:04:05")))
	return nil
}
