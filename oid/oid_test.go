package oid_test

import (
	"os"
	"testing"
	"time"

	"github.com/acudac-com/public-go/oid"
)

var client = oid.NewClient("http://localhost:18090", os.Getenv("ACUDAC_IDENTITY_CLIENT_ID"), os.Getenv("ACUDAC_IDENTITY_CLIENT_SECRET"))

func init() {
	oid.AddClient(client)
}

func Test_All(t *testing.T) {
	code := "LNDJF7X5XG5BIEBNJYHDRWIN3V"
	redirectURL := "http://localhost:18090/callback"
	tokens, err := client.ExchangeCode(&code, &redirectURL)
	if err != nil {
		t.Fatal(err)
	}

	// should not refresh
	identity, err := oid.Authenticate(time.Now(), &tokens.IDToken, &tokens.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if identity.Refreshed() {
		t.Fatal("identity should not be refreshed")
	}

	// should refresh
	identity, err = oid.Authenticate(time.Now().Add(time.Hour*24), &tokens.IDToken, &tokens.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if !identity.Refreshed() {
		t.Fatal("identity should be refreshed")
	}
}
