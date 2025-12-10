// Package oid can be used to authenticate OIDC ID tokens. At this stage only ED25519 algorithm is supported.
package oid

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/acudac-com/public-go/rest"
)

// Issuer is an OIDC issuer, like Acudac Identity
type Issuer struct {
	URL                   string    // e.g. https://identity.acudac.com
	JwksURL               string    // e.g. https://identity.acudac.com/.well-known/jwks.json
	TokensURL             string    // e.g. https://identity.acudac.com/token
	Clients               []*Client // at least one client that was created at the issuer
	pubicKeys             map[string]*ed25519.PublicKey
	publicKeysMu          *sync.RWMutex
	publicKeysLastFetched time.Time
}

// Client is an OIDC client created at an OIDC issuer.
type Client struct {
	ID     string
	Secret string
	issuer *Issuer
}

// Refresh refreshes the tokens with the given refresh token.
func (c *Client) Refresh(refreshToken *string, idToken *string) error {
	restClient := rest.NewClient(http.DefaultClient, c.issuer.TokensURL)
	form := url.Values{
		"client_id":     {c.ID},
		"grant_type":    {"refresh_token"},
		"refresh_token": {*refreshToken},
		"client_secret": {c.Secret},
	}
	type Tokens struct {
		AccessToken  string `json:"access_token"`
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
	}
	tokens := &Tokens{}
	if err := restClient.PostForm("", form, tokens); err != nil {
		return err
	}
	*refreshToken = tokens.RefreshToken
	*idToken = tokens.IDToken
	return nil
}

var issuers = map[string]*Issuer{}

// AddIssuer adds the given issuer. Not concurrency proof so use in only one init function.
func AddIssuer(issuer *Issuer) error {
	if issuer.URL == "" {
		return fmt.Errorf("issuer.URL cannot be empty")
	}
	if issuer.JwksURL == "" {
		return fmt.Errorf("issuer.JwksURL cannot be empty")
	}
	if len(issuer.Clients) == 0 {
		return fmt.Errorf("issuer.Clients cannot be empty")
	}

	// validate clients and add issuer to each
	for i, client := range issuer.Clients {
		if client.ID == "" {
			return fmt.Errorf("issuer.Clients[%d].ID cannot be empty", i)
		}
		if client.Secret == "" {
			return fmt.Errorf("issuer.Clients[%d].Secret cannot be empty", i)
		}
		client.issuer = issuer
	}

	// setup private vars
	issuer.publicKeysMu = &sync.RWMutex{}
	issuer.pubicKeys = map[string]*ed25519.PublicKey{}

	// add to global list of issuers
	issuers[issuer.URL] = issuer
	return nil
}

// Identity is an OIDC identity extracted from a valid ID token
type Identity struct {
	Sub       string // the ID of the user/machine
	Email     string
	Aud       string // the client_id
	Iat       int64  // when the token was issued
	Exp       int64  // when the token expires
	Iss       string // the issuer of the token
	refreshed bool   // whether the tokens were refreshed
}

// Refreshed returns whether the tokens given to the [Authenticate] were refreshed.
// If true, you must save the updated tokens for future use, e.g. in the user's cookies.
func (i *Identity) Refreshed() bool {
	return i.refreshed
}

// Authenticate returns the verified identity of the given ID token.
// If a refresh token is provided, it will automatically refresh the ID token if it expired.
func Authenticate(now time.Time, idToken *string, refreshToken *string) (*Identity, error) {
	if idToken == nil {
		return nil, fmt.Errorf("id token cannot be nil")
	}

	// split id token into header, body and signature
	idTokenParts := strings.Split(*idToken, ".")
	if len(idTokenParts) != 3 {
		return nil, fmt.Errorf("id token must have format <header>.<body>.<signature>")
	}
	headerString, body, signature := idTokenParts[0], idTokenParts[1], idTokenParts[2]

	// parse and validate header
	type Header struct {
		Kid string // e.g. 194md12x
		Alg string // e.g. EdDSA
		Typ string // must be JWT
	}
	header := &Header{}
	if err := unmarshalB64(headerString, header); err != nil {
		return nil, fmt.Errorf("parsing header: %w", err)
	}
	if header.Alg != "EdDSA" {
		return nil, fmt.Errorf("unsupported header.alg: %s", header.Alg)
	}
	if header.Typ != "JWT" {
		return nil, fmt.Errorf("unsupported header.typ: %s", header.Typ)
	}

	// extract identity from body
	identity := &Identity{}
	if err := unmarshalB64(body, identity); err != nil {
		return nil, fmt.Errorf("parsing header: %w", err)
	}

	// fail if id token is invalid
	if identity.Iss == "" {
		return nil, fmt.Errorf("id token missing 'iss'")
	}
	if identity.Aud == "" {
		return nil, fmt.Errorf("id token missing 'aud'")
	}

	// find issuer
	issuer, ok := issuers[identity.Iss]
	if !ok {
		return nil, fmt.Errorf("%s is not an accepted id token issuer", identity.Iss)
	}

	// find client with same id as the ID token audience
	client, err := issuer.Client(identity.Aud)
	if err != nil {
		return nil, err
	}

	// try to refresh id token if expired
	if time.Unix(identity.Exp, 0).Before(now) {
		if refreshToken == nil || issuer.TokensURL == "" {
			return nil, fmt.Errorf("id token expired but no refresh token provided")
		}
		if err := client.Refresh(refreshToken, idToken); err != nil {
			return nil, fmt.Errorf("refreshing tokens: %v", err)
		}
		identity.refreshed = true
	}

	// validate the signature
	if err := issuer.ValidateSignature(header.Kid, headerString+"."+body, signature); err != nil {
		return nil, err
	}

	return identity, nil
}

// unmarshalB64 first base64 decodes the url encoded string, then JSON unmarshals the decoded bytes into the given obj.
func unmarshalB64(value string, obj any) error {
	// decode base64 url encoding
	headerBytes, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return fmt.Errorf("base64 decoding: %w", err)
	}

	// json unmarshal
	if err := json.Unmarshal(headerBytes, obj); err != nil {
		return fmt.Errorf("json unmarshalling header: %w", err)
	}
	return nil
}

// Client returns the client with the given id
func (i *Issuer) Client(id string) (*Client, error) {
	var client *Client
	for _, issuerClient := range i.Clients {
		if issuerClient.ID == id {
			client = issuerClient
			break
		}
	}
	if client == nil {
		return nil, fmt.Errorf("no client with id=%s", id)
	}
	return client, nil
}

func (i *Issuer) ValidateSignature(kid string, signingInput string, signature string) error {
	publicKey, err := i.PublicKey(time.Now(), kid)
	if err != nil {
		return err
	}
	decodedSignature, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("base64 decoding signature: %w", err)
	}
	if len(decodedSignature) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signature length: expected %d, got %d", ed25519.SignatureSize, len(decodedSignature))
	}
	if !ed25519.Verify(*publicKey, []byte(signingInput), decodedSignature) {
		return errors.New("signature does not match")
	}
	return nil
}

// Refresh the tokens with the given refresh token

// PublicKey returns the public key with the given id.
// It refreshes the list of public keys from the issuer's JWKS url every hour.
func (i *Issuer) PublicKey(now time.Time, kid string) (*ed25519.PublicKey, error) {
	// if fetched less than an hour aga, read from cache
	if i.publicKeysLastFetched.Add(1 * time.Hour).After(now) {
		return i.publicKeyFromCache(kid)
	}

	// lock and defer unlock
	i.publicKeysMu.Lock()
	defer i.publicKeysMu.Unlock()

	// check last fetched time again incase it has been updated while waiting for lock
	if i.publicKeysLastFetched.Add(1 * time.Hour).After(now) {
		return i.publicKeyFromCache(kid)
	}

	// fetch jwks
	jwks, err := i.Jwks()
	if err != nil {
		return nil, err
	}

	// convert jwks to public keys
	publicKeys, err := PublicKeysFromJwks(jwks)
	if err != nil {
		return nil, fmt.Errorf("converting jwks into public keys: %w", err)
	}

	// save to cache and return requested public key from cache
	i.pubicKeys = publicKeys
	i.publicKeysLastFetched = now
	return i.publicKeyFromCache(kid)
}

func (i *Issuer) publicKeyFromCache(kid string) (*ed25519.PublicKey, error) {
	if publicKey, ok := i.pubicKeys[kid]; ok {
		return publicKey, nil
	}
	return nil, fmt.Errorf("no public key found with kid=%s", kid)
}

type JWKS struct {
	Keys []*JWK `json:"keys"`
}
type JWK struct {
	Kty string `json:"kty"` // e.g. OKP
	Crv string `json:"crv"` // e.g. Ed25519
	Alg string `json:"alg"` // e.g. EdDSA
	Use string `json:"use"` // e.g. sig
	Kid string `json:"kid"` // e.g. 194md4sb
	X   string `json:"x"`   // e.g. asdf98qh4rpoqierqp98asc9as-asdfhsdfahsd98
}

// Jwks fetches the issuer's JWKS from its JWKS url.
func (i *Issuer) Jwks() (*JWKS, error) {
	restClient := rest.NewClient(http.DefaultClient, i.JwksURL)
	jwks := &JWKS{}
	if err := restClient.Get("", jwks); err != nil {
		return nil, fmt.Errorf("fetching JWKS from %s: %w", i.JwksURL, err)
	}
	return jwks, nil
}

// PublicKeysFromJwks converts the given JWKS to ed25519 public keys
func PublicKeysFromJwks(jwks *JWKS) (map[string]*ed25519.PublicKey, error) {
	publicKeysMap := map[string]*ed25519.PublicKey{}
	for _, jwk := range jwks.Keys {
		publicKey, err := PublicKeyFromJwk(jwk)
		if err != nil {
			return nil, err
		}
		publicKeysMap[jwk.Kid] = publicKey
	}
	return publicKeysMap, nil
}

// PublicKeyFromJwk converts the given JWK to an ed25519 public key
func PublicKeyFromJwk(jwk *JWK) (*ed25519.PublicKey, error) {
	if jwk.Kty != "OKP" {
		return nil, fmt.Errorf("unsupported key type: %s", jwk.Kty)
	}
	if jwk.Crv != "Ed25519" {
		return nil, fmt.Errorf("unsupported OKP curve: %s", jwk.Crv)
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("invalid OKP X public key: %w", err)
	}
	if len(xBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid OKP X public key length: expected %d, got %d", ed25519.PublicKeySize, len(xBytes))
	}
	ed25519PublicKey := ed25519.PublicKey(xBytes)
	return &ed25519PublicKey, nil
}
