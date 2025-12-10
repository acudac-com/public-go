package oid_test

import (
	"os"
	"testing"
	"time"

	"github.com/acudac-com/public-go/oid"
)

func init() {
	oid.AddIssuer(&oid.Issuer{
		URL:       "http://localhost:18090",
		JwksURL:   "http://localhost:18090/.well-known/jwks.json",
		TokensURL: "http://localhost:18090/token",
		Clients:   []*oid.Client{{ID: os.Getenv("ACUDAC_IDENTITY_CLIENT_ID"), Secret: os.Getenv("ACUDAC_IDENTITY_CLIENT_SECRET")}},
	})
}

func Test_Authenticate(t *testing.T) {
	idToken := os.Getenv("ACUDAC_IDENTITY_ID_TOKEN")
	refreshToken := os.Getenv("ACUDAC_IDENTITY_REFRESH_TOKEN")
	identity, err := oid.Authenticate(time.Now(), &idToken, &refreshToken)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", identity)
	identity, err = oid.Authenticate(time.Now(), &idToken, &refreshToken)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", identity)
}
