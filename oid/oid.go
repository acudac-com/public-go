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

// Client is a client of an OIDC issuer like Acudac Identity
type Client struct {
	IssuerURL             string // e.g. https://identity.acudac.com
	JwksURL               string // e.g. https://identity.acudac.com/.well-known/jwks.json
	TokensURL             string // e.g. https://identity.acudac.com/token
	ID                    string
	Secret                string
	pubicKeys             map[string]*ed25519.PublicKey
	publicKeysMu          *sync.RWMutex
	publicKeysLastFetched time.Time
}

// NewClient returns a new client for the given issuer URL, client ID, and client secret.
// By default JwksURL is set to the issuer URL with /.well-known/jwks.json appended.
// By default TokensURL is set to the issuer URL with /token appended.
func NewClient(issuerURL, id, secret string) *Client {
	return &Client{
		IssuerURL: issuerURL,
		JwksURL:   issuerURL + "/.well-known/jwks.json",
		TokensURL: issuerURL + "/token",
		ID:        id,
		Secret:    secret,
	}
}

// Tokens is a set of tokens returned from an issuer's token endpoint
type Tokens struct {
	AccessToken  string `json:"access_token,omitempty"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
}

// ExchangeCode exchanges the given code for tokens.
func (c *Client) ExchangeCode(code *string, redirectURL *string) (*Tokens, error) {
	restClient := rest.NewClient(http.DefaultClient, c.TokensURL)
	form := url.Values{
		"client_id":     {c.ID},
		"grant_type":    {"authorization_code"},
		"code":          {*code},
		"redirect_url":  {*redirectURL},
		"client_secret": {c.Secret},
	}
	tokens := &Tokens{}
	if err := restClient.PostForm("", form, tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

// Refresh refreshes the tokens with the given refresh token.
func (c *Client) Refresh(refreshToken *string, idToken *string) error {
	restClient := rest.NewClient(http.DefaultClient, c.TokensURL)
	form := url.Values{
		"client_id":     {c.ID},
		"grant_type":    {"refresh_token"},
		"refresh_token": {*refreshToken},
		"client_secret": {c.Secret},
	}
	tokens := &Tokens{}
	if err := restClient.PostForm("", form, tokens); err != nil {
		return err
	}
	*refreshToken = tokens.RefreshToken
	*idToken = tokens.IDToken
	return nil
}

var clients = map[string]*Client{}

// AddClient adds the given client so that the [Authenticate] function can use it.
// Not concurrency proof so use in only one init function.
func AddClient(client *Client) error {
	if client.IssuerURL == "" {
		return fmt.Errorf("client.IssuerURL cannot be empty")
	}
	if client.JwksURL == "" {
		return fmt.Errorf("client.JwksURL cannot be empty")
	}
	if client.ID == "" {
		return fmt.Errorf("client.ID cannot be empty")
	}
	// issuer secret is optional

	// setup private vars
	client.publicKeysMu = &sync.RWMutex{}
	client.pubicKeys = map[string]*ed25519.PublicKey{}

	// add to global list of issuers
	clients[client.IssuerURL] = client
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

	// parse jwt
	jwt, err := ParseJWT(idToken)
	if err != nil {
		return nil, err
	}

	// find issuer
	issuer, ok := clients[jwt.Identity.Iss]
	if !ok {
		return nil, fmt.Errorf("%s is not an accepted id token issuer", jwt.Identity.Iss)
	}

	// try to refresh id token if expired
	if time.Unix(jwt.Identity.Exp, 0).Before(now) {
		if refreshToken == nil || issuer.TokensURL == "" {
			return nil, fmt.Errorf("id token expired but no refresh token provided")
		}
		if err := issuer.Refresh(refreshToken, idToken); err != nil {
			return nil, fmt.Errorf("refreshing tokens: %v", err)
		}

		// re-parse jwt, since idToken has been refreshed
		jwt, err = ParseJWT(idToken)
		if err != nil {
			return nil, err
		}
		jwt.Identity.refreshed = true
	}

	// validate the signature
	if err := issuer.ValidateSignature(jwt.Header.Kid, jwt.SignedString, jwt.Signature); err != nil {
		return nil, err
	}

	return jwt.Identity, nil
}

// JWT is a parsed JWT
type JWT struct {
	Header       *Header
	Identity     *Identity
	SignedString string
	Signature    string
}

// Header is a JWT header
type Header struct {
	Kid string // e.g. 194md12x
	Alg string // e.g. EdDSA
	Typ string // must be JWT
}

// ParseJWT parses the given id token into its header, body and signature.
func ParseJWT(idToken *string) (*JWT, error) {
	// split id token into header, body and signature
	idTokenParts := strings.Split(*idToken, ".")
	if len(idTokenParts) != 3 {
		return nil, fmt.Errorf("id token must have format <header>.<body>.<signature>")
	}
	headerString, body, signature := idTokenParts[0], idTokenParts[1], idTokenParts[2]

	// parse and validate header
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

	// validate issuer and audience
	if identity.Iss == "" {
		return nil, fmt.Errorf("id token missing 'iss'")
	}
	if identity.Aud == "" {
		return nil, fmt.Errorf("id token missing 'aud'")
	}

	// return jwt
	return &JWT{
		Header:       header,
		Identity:     identity,
		SignedString: headerString + "." + body,
		Signature:    signature,
	}, nil
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

func (c *Client) ValidateSignature(kid string, signingInput string, signature string) error {
	publicKey, err := c.PublicKey(time.Now(), kid)
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

// PublicKey returns the public key with the given id.
// It refreshes the list of public keys from the issuer's JWKS url every hour.
func (c *Client) PublicKey(now time.Time, kid string) (*ed25519.PublicKey, error) {
	// if fetched less than an hour aga, read from cache
	if c.publicKeysLastFetched.Add(1 * time.Hour).After(now) {
		return c.publicKeyFromCache(kid)
	}

	// lock and defer unlock
	c.publicKeysMu.Lock()
	defer c.publicKeysMu.Unlock()

	// check last fetched time again incase it has been updated while waiting for lock
	if c.publicKeysLastFetched.Add(1 * time.Hour).After(now) {
		return c.publicKeyFromCache(kid)
	}

	// fetch jwks
	jwks, err := c.Jwks()
	if err != nil {
		return nil, err
	}

	// convert jwks to public keys
	publicKeys, err := PublicKeysFromJwks(jwks)
	if err != nil {
		return nil, fmt.Errorf("converting jwks into public keys: %w", err)
	}

	// save to cache and return requested public key from cache
	c.pubicKeys = publicKeys
	c.publicKeysLastFetched = now
	return c.publicKeyFromCache(kid)
}

func (c *Client) publicKeyFromCache(kid string) (*ed25519.PublicKey, error) {
	if publicKey, ok := c.pubicKeys[kid]; ok {
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
func (c *Client) Jwks() (*JWKS, error) {
	restClient := rest.NewClient(http.DefaultClient, c.JwksURL)
	jwks := &JWKS{}
	if err := restClient.Get("", jwks); err != nil {
		return nil, fmt.Errorf("fetching JWKS from %s: %w", c.JwksURL, err)
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
