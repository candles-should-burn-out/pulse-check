package internal

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	testIssuer    = "http://keycloak.test/realms/pulse-check"
	testAudience  = "pulse-check-api"
	testKeyID     = "test-key"
	testSubjectID = "11111111-1111-4111-8111-111111111111"
)

func TestAuthMiddlewareRejectsMissingToken(t *testing.T) {
	authenticator := newTestAuthenticator(t)
	app := NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	app.authenticator = authenticator.authenticator

	recorder := httptest.NewRecorder()
	app.Routes().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/entities", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(recorder.Body.String(), "missing_bearer_token") {
		t.Fatalf("body = %q, want missing_bearer_token", recorder.Body.String())
	}
}

func TestHealthEndpointsRemainPublicWithAuthConfigured(t *testing.T) {
	authenticator := newTestAuthenticator(t)
	app := NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)))
	app.authenticator = authenticator.authenticator

	recorder := httptest.NewRecorder()
	app.Routes().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestAuthenticatorAcceptsValidToken(t *testing.T) {
	authenticator := newTestAuthenticator(t)
	token := authenticator.signToken(t, tokenClaims{
		Issuer:         testIssuer,
		Audience:       []string{testAudience},
		ExpirationTime: authenticator.now.Add(time.Hour).Unix(),
	})

	claims, err := authenticator.authenticator.ValidateBearerToken(context.Background(), "Bearer "+token)
	if err != nil {
		t.Fatalf("ValidateBearerToken() error = %v", err)
	}

	if claims.Subject != testSubjectID {
		t.Fatalf("Subject = %q, want %s", claims.Subject, testSubjectID)
	}
}

func TestAuthenticatorRejectsNonUUIDSubject(t *testing.T) {
	authenticator := newTestAuthenticator(t)
	token := authenticator.signToken(t, tokenClaims{
		Subject:        "user-1",
		Issuer:         testIssuer,
		Audience:       []string{testAudience},
		ExpirationTime: authenticator.now.Add(time.Hour).Unix(),
	})

	_, err := authenticator.authenticator.ValidateBearerToken(context.Background(), "Bearer "+token)
	if err == nil || authErrorCode(err) != "invalid_subject" {
		t.Fatalf("ValidateBearerToken() error = %v, want invalid subject", err)
	}
}

func TestAuthenticatorRejectsMalformedToken(t *testing.T) {
	authenticator := newTestAuthenticator(t)

	_, err := authenticator.authenticator.ValidateBearerToken(context.Background(), "Bearer not-a-jwt")
	if err == nil || authErrorCode(err) != "invalid_token" {
		t.Fatalf("ValidateBearerToken() error = %v, want invalid token", err)
	}
}

func TestAuthenticatorRejectsInvalidSignature(t *testing.T) {
	authenticator := newTestAuthenticator(t)
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate other key: %v", err)
	}

	token := signToken(t, otherKey, testKeyID, tokenClaims{
		Issuer:         testIssuer,
		Audience:       []string{testAudience},
		ExpirationTime: authenticator.now.Add(time.Hour).Unix(),
	})

	_, err = authenticator.authenticator.ValidateBearerToken(context.Background(), "Bearer "+token)
	if err == nil || !strings.Contains(authErrorCode(err), "invalid_token") {
		t.Fatalf("ValidateBearerToken() error = %v, want invalid token", err)
	}
}

func TestAuthenticatorRejectsWrongIssuer(t *testing.T) {
	authenticator := newTestAuthenticator(t)
	token := authenticator.signToken(t, tokenClaims{
		Issuer:         "http://keycloak.test/realms/other",
		Audience:       []string{testAudience},
		ExpirationTime: authenticator.now.Add(time.Hour).Unix(),
	})

	_, err := authenticator.authenticator.ValidateBearerToken(context.Background(), "Bearer "+token)
	if err == nil || authErrorCode(err) != "invalid_issuer" {
		t.Fatalf("ValidateBearerToken() error = %v, want invalid issuer", err)
	}
}

func TestAuthenticatorRejectsWrongAudience(t *testing.T) {
	authenticator := newTestAuthenticator(t)
	token := authenticator.signToken(t, tokenClaims{
		Issuer:         testIssuer,
		Audience:       []string{"other-api"},
		ExpirationTime: authenticator.now.Add(time.Hour).Unix(),
	})

	_, err := authenticator.authenticator.ValidateBearerToken(context.Background(), "Bearer "+token)
	if err == nil || authErrorCode(err) != "invalid_audience" {
		t.Fatalf("ValidateBearerToken() error = %v, want invalid audience", err)
	}
}

func TestAuthenticatorRejectsExpiredToken(t *testing.T) {
	authenticator := newTestAuthenticator(t)
	token := authenticator.signToken(t, tokenClaims{
		Issuer:         testIssuer,
		Audience:       []string{testAudience},
		ExpirationTime: authenticator.now.Add(-time.Minute).Unix(),
	})

	_, err := authenticator.authenticator.ValidateBearerToken(context.Background(), "Bearer "+token)
	if err == nil || authErrorCode(err) != "expired_token" {
		t.Fatalf("ValidateBearerToken() error = %v, want expired token", err)
	}
}

func TestAuthenticatorRejectsTokenNotValidYet(t *testing.T) {
	authenticator := newTestAuthenticator(t)
	notBefore := authenticator.now.Add(time.Minute).Unix()
	token := authenticator.signToken(t, tokenClaims{
		Issuer:         testIssuer,
		Audience:       []string{testAudience},
		ExpirationTime: authenticator.now.Add(time.Hour).Unix(),
		NotBefore:      &notBefore,
	})

	_, err := authenticator.authenticator.ValidateBearerToken(context.Background(), "Bearer "+token)
	if err == nil || authErrorCode(err) != "token_not_valid_yet" {
		t.Fatalf("ValidateBearerToken() error = %v, want token not valid yet", err)
	}
}

type testAuthenticator struct {
	authenticator *Authenticator
	privateKey    *rsa.PrivateKey
	now           time.Time
}

type tokenClaims struct {
	Subject        string   `json:"sub"`
	Issuer         string   `json:"iss"`
	Audience       []string `json:"aud"`
	ExpirationTime int64    `json:"exp"`
	NotBefore      *int64   `json:"nbf,omitempty"`
}

func newTestAuthenticator(t *testing.T) testAuthenticator {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	authenticator, err := NewAuthenticator(AuthConfig{
		Issuer:   testIssuer,
		JWKSURL:  "https://keycloak.test/realms/pulse-check/protocol/openid-connect/certs",
		Audience: testAudience,
	})
	if err != nil {
		t.Fatalf("NewAuthenticator() error = %v", err)
	}

	now := time.Unix(1_700_000_000, 0)
	authenticator.now = func() time.Time { return now }
	authenticator.verifier = oidc.NewVerifier(
		testIssuer,
		&oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{&privateKey.PublicKey}},
		&oidc.Config{
			SkipClientIDCheck:    true,
			Now:                  authenticator.now,
			SupportedSigningAlgs: []string{oidc.RS256},
		},
	)

	return testAuthenticator{
		authenticator: authenticator,
		privateKey:    privateKey,
		now:           now,
	}
}

func (auth testAuthenticator) signToken(t *testing.T, claims tokenClaims) string {
	t.Helper()

	if claims.Subject == "" {
		claims.Subject = testSubjectID
	}

	return signToken(t, auth.privateKey, testKeyID, claims)
}

func signToken(t *testing.T, privateKey *rsa.PrivateKey, keyID string, claims tokenClaims) string {
	t.Helper()

	header := map[string]string{
		"alg": "RS256",
		"kid": keyID,
		"typ": "JWT",
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsBytes)
	signedData := encodedHeader + "." + encodedClaims
	digest := sha256.Sum256([]byte(signedData))

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return signedData + "." + base64.RawURLEncoding.EncodeToString(signature)
}
