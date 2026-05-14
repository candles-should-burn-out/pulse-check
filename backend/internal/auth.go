package internal

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

const authContextKey contextKey = "auth_claims"

var (
	errAuthDisabled      = errors.New("auth disabled")
	errMissingBearer     = errors.New("missing bearer token")
	errInvalidToken      = errors.New("invalid token")
	errInvalidIssuer     = errors.New("invalid issuer")
	errInvalidAudience   = errors.New("invalid audience")
	errExpiredToken      = errors.New("expired token")
	errTokenNotValidYet  = errors.New("token not valid yet")
	errMissingRole       = errors.New("missing required role")
	errUnsupportedKeyAlg = errors.New("unsupported key algorithm")
)

type contextKey string

type Authenticator struct {
	issuer       string
	jwksURL      string
	audience     string
	requiredRole string
	httpClient   *http.Client
	now          func() time.Time

	mu       sync.RWMutex
	keyCache map[string]*rsa.PublicKey
}

type TokenClaims struct {
	Subject        string
	Issuer         string
	Audience       []string
	ExpirationTime time.Time
	NotBefore      *time.Time
	Roles          map[string]struct{}
}

type jwtClaims struct {
	Subject        string                `json:"sub"`
	Issuer         string                `json:"iss"`
	Audience       audienceClaim         `json:"aud"`
	ExpirationTime int64                 `json:"exp"`
	NotBefore      *int64                `json:"nbf,omitempty"`
	RealmAccess    rolesClaim            `json:"realm_access,omitempty"`
	ResourceAccess map[string]rolesClaim `json:"resource_access,omitempty"`
}

type rolesClaim struct {
	Roles []string `json:"roles"`
}

type audienceClaim []string

type jwksDocument struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KeyID     string `json:"kid"`
	KeyType   string `json:"kty"`
	Algorithm string `json:"alg,omitempty"`
	Use       string `json:"use,omitempty"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

func NewAuthenticator(config AuthConfig) (*Authenticator, error) {
	if config.Issuer == "" && config.JWKSURL == "" && config.Audience == "" {
		return nil, errAuthDisabled
	}

	if config.Issuer == "" || config.JWKSURL == "" || config.Audience == "" {
		return nil, fmt.Errorf("OIDC_ISSUER, OIDC_JWKS_URL, and OIDC_AUDIENCE must be set together")
	}

	return &Authenticator{
		issuer:       strings.TrimRight(config.Issuer, "/"),
		jwksURL:      config.JWKSURL,
		audience:     config.Audience,
		requiredRole: config.RequiredRole,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		now:      time.Now,
		keyCache: map[string]*rsa.PublicKey{},
	}, nil
}

func AuthClaims(ctx context.Context) (*TokenClaims, bool) {
	claims, ok := ctx.Value(authContextKey).(*TokenClaims)
	return claims, ok
}

func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.ValidateBearerToken(r.Context(), r.Header.Get("Authorization"))
		if err != nil {
			status := http.StatusUnauthorized
			if errors.Is(err, errMissingRole) {
				status = http.StatusForbidden
			}

			w.Header().Set("WWW-Authenticate", "Bearer")
			respondJSON(w, status, map[string]string{"error": authErrorCode(err)})
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), authContextKey, claims)))
	})
}

func (a *Authenticator) ValidateBearerToken(ctx context.Context, authorization string) (*TokenClaims, error) {
	token, ok := strings.CutPrefix(authorization, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return nil, errMissingBearer
	}

	return a.ValidateToken(ctx, strings.TrimSpace(token))
}

func (a *Authenticator) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errInvalidToken
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errInvalidToken
	}

	var header struct {
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, errInvalidToken
	}
	if header.Algorithm != "RS256" || header.KeyID == "" {
		return nil, errInvalidToken
	}

	publicKey, err := a.publicKey(ctx, header.KeyID)
	if err != nil {
		return nil, err
	}

	signedData := parts[0] + "." + parts[1]
	digest := sha256.Sum256([]byte(signedData))
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errInvalidToken
	}
	if err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature); err != nil {
		return nil, errInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errInvalidToken
	}

	var rawClaims jwtClaims
	if err := json.Unmarshal(payloadBytes, &rawClaims); err != nil {
		return nil, errInvalidToken
	}

	claims := TokenClaims{
		Subject:        rawClaims.Subject,
		Issuer:         strings.TrimRight(rawClaims.Issuer, "/"),
		Audience:       []string(rawClaims.Audience),
		ExpirationTime: time.Unix(rawClaims.ExpirationTime, 0),
		Roles:          rawClaims.roles(),
	}

	if rawClaims.NotBefore != nil {
		notBefore := time.Unix(*rawClaims.NotBefore, 0)
		claims.NotBefore = &notBefore
	}

	if claims.Issuer != a.issuer {
		return nil, errInvalidIssuer
	}
	if !claims.hasAudience(a.audience) {
		return nil, errInvalidAudience
	}

	now := a.now()
	if !claims.ExpirationTime.After(now) {
		return nil, errExpiredToken
	}
	if claims.NotBefore != nil && claims.NotBefore.After(now) {
		return nil, errTokenNotValidYet
	}
	if a.requiredRole != "" && !claims.hasRole(a.requiredRole) {
		return nil, errMissingRole
	}

	return &claims, nil
}

func (a *Authenticator) publicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	a.mu.RLock()
	publicKey := a.keyCache[keyID]
	a.mu.RUnlock()
	if publicKey != nil {
		return publicKey, nil
	}

	if err := a.refreshKeys(ctx); err != nil {
		return nil, err
	}

	a.mu.RLock()
	publicKey = a.keyCache[keyID]
	a.mu.RUnlock()
	if publicKey == nil {
		return nil, errInvalidToken
	}

	return publicKey, nil
}

func (a *Authenticator) refreshKeys(ctx context.Context) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, a.jwksURL, nil)
	if err != nil {
		return errInvalidToken
	}

	response, err := a.httpClient.Do(request)
	if err != nil {
		return errInvalidToken
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errInvalidToken
	}

	var document jwksDocument
	if err := json.NewDecoder(response.Body).Decode(&document); err != nil {
		return errInvalidToken
	}

	nextCache := make(map[string]*rsa.PublicKey, len(document.Keys))
	for _, key := range document.Keys {
		publicKey, err := key.publicKey()
		if err != nil {
			if errors.Is(err, errUnsupportedKeyAlg) {
				continue
			}
			return errInvalidToken
		}
		nextCache[key.KeyID] = publicKey
	}

	a.mu.Lock()
	a.keyCache = nextCache
	a.mu.Unlock()

	return nil
}

func (key jwk) publicKey() (*rsa.PublicKey, error) {
	if key.KeyID == "" || key.KeyType != "RSA" || key.Modulus == "" || key.Exponent == "" {
		return nil, errUnsupportedKeyAlg
	}
	if key.Algorithm != "" && key.Algorithm != "RS256" {
		return nil, errUnsupportedKeyAlg
	}
	if key.Use != "" && key.Use != "sig" {
		return nil, errUnsupportedKeyAlg
	}

	modulus, err := base64.RawURLEncoding.DecodeString(key.Modulus)
	if err != nil {
		return nil, err
	}
	exponentBytes, err := base64.RawURLEncoding.DecodeString(key.Exponent)
	if err != nil {
		return nil, err
	}

	exponent := new(big.Int).SetBytes(exponentBytes)
	if !exponent.IsInt64() {
		return nil, errUnsupportedKeyAlg
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(modulus),
		E: int(exponent.Int64()),
	}, nil
}

func (claims jwtClaims) roles() map[string]struct{} {
	roles := map[string]struct{}{}
	for _, role := range claims.RealmAccess.Roles {
		roles[role] = struct{}{}
	}
	for _, access := range claims.ResourceAccess {
		for _, role := range access.Roles {
			roles[role] = struct{}{}
		}
	}
	return roles
}

func (claims TokenClaims) hasAudience(audience string) bool {
	for _, value := range claims.Audience {
		if value == audience {
			return true
		}
	}
	return false
}

func (claims TokenClaims) hasRole(role string) bool {
	_, ok := claims.Roles[role]
	return ok
}

func (audience *audienceClaim) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*audience = []string{single}
		return nil
	}

	var multiple []string
	if err := json.Unmarshal(data, &multiple); err != nil {
		return err
	}

	*audience = multiple
	return nil
}

func authErrorCode(err error) string {
	switch {
	case errors.Is(err, errMissingBearer):
		return "missing_bearer_token"
	case errors.Is(err, errInvalidIssuer):
		return "invalid_issuer"
	case errors.Is(err, errInvalidAudience):
		return "invalid_audience"
	case errors.Is(err, errExpiredToken):
		return "expired_token"
	case errors.Is(err, errTokenNotValidYet):
		return "token_not_valid_yet"
	case errors.Is(err, errMissingRole):
		return "missing_required_role"
	default:
		return "invalid_token"
	}
}
