package internal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
)

const authContextKey contextKey = "auth_claims"

var (
	errAuthDisabled     = errors.New("auth disabled")
	errMissingBearer    = errors.New("missing bearer token")
	errInvalidToken     = errors.New("invalid token")
	errInvalidIssuer    = errors.New("invalid issuer")
	errInvalidAudience  = errors.New("invalid audience")
	errInvalidSubject   = errors.New("invalid subject")
	errExpiredToken     = errors.New("expired token")
	errTokenNotValidYet = errors.New("token not valid yet")
	errMissingRole      = errors.New("missing required role")
)

type contextKey string

type Authenticator struct {
	issuer       string
	audience     string
	requiredRole string
	now          func() time.Time
	verifier     tokenVerifier
}

type tokenVerifier interface {
	Verify(context.Context, string) (*oidc.IDToken, error)
}

type TokenClaims struct {
	Subject        string
	Issuer         string
	Audience       []string
	ExpirationTime time.Time
	NotBefore      *time.Time
	Roles          map[string]struct{}
}

type keycloakClaims struct {
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

func NewAuthenticator(config AuthConfig) (*Authenticator, error) {
	if config.Issuer == "" && config.JWKSURL == "" && config.Audience == "" {
		return nil, errAuthDisabled
	}

	if config.Issuer == "" || config.JWKSURL == "" || config.Audience == "" {
		return nil, fmt.Errorf("OIDC_ISSUER, OIDC_JWKS_URL, and OIDC_AUDIENCE must be set together")
	}

	issuer := strings.TrimRight(config.Issuer, "/")
	now := time.Now
	keySet := oidc.NewRemoteKeySet(context.Background(), config.JWKSURL)
	verifier := oidc.NewVerifier(issuer, keySet, &oidc.Config{
		SkipClientIDCheck:    true,
		Now:                  now,
		SupportedSigningAlgs: []string{oidc.RS256},
	})

	return &Authenticator{
		issuer:       issuer,
		audience:     config.Audience,
		requiredRole: config.RequiredRole,
		now:          now,
		verifier:     verifier,
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
	rawClaims, parseErr := parseTokenClaims(token)
	if parseErr != nil {
		return nil, parseErr
	}

	idToken, err := a.verifier.Verify(ctx, token)
	if err != nil {
		return nil, a.classifyVerifyError(rawClaims)
	}

	if err := idToken.Claims(&rawClaims); err != nil {
		return nil, errInvalidToken
	}

	claims := rawClaims.toTokenClaims()
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, errInvalidSubject
	}
	if !claims.hasAudience(a.audience) {
		return nil, errInvalidAudience
	}
	if claims.NotBefore != nil && claims.NotBefore.After(a.now()) {
		return nil, errTokenNotValidYet
	}
	if a.requiredRole != "" && !claims.hasRole(a.requiredRole) {
		return nil, errMissingRole
	}

	return claims, nil
}

func (a *Authenticator) classifyVerifyError(claims keycloakClaims) error {
	now := a.now()
	if strings.TrimRight(claims.Issuer, "/") != a.issuer {
		return errInvalidIssuer
	}
	if !claims.Audience.has(a.audience) {
		return errInvalidAudience
	}
	if claims.ExpirationTime != 0 && !time.Unix(claims.ExpirationTime, 0).After(now) {
		return errExpiredToken
	}
	if claims.NotBefore != nil && time.Unix(*claims.NotBefore, 0).After(now) {
		return errTokenNotValidYet
	}

	return errInvalidToken
}

func (claims keycloakClaims) toTokenClaims() *TokenClaims {
	tokenClaims := &TokenClaims{
		Subject:        claims.Subject,
		Issuer:         strings.TrimRight(claims.Issuer, "/"),
		Audience:       []string(claims.Audience),
		ExpirationTime: time.Unix(claims.ExpirationTime, 0),
		Roles:          claims.roles(),
	}

	if claims.NotBefore != nil {
		notBefore := time.Unix(*claims.NotBefore, 0)
		tokenClaims.NotBefore = &notBefore
	}

	return tokenClaims
}

func parseTokenClaims(token string) (keycloakClaims, error) {
	var claims keycloakClaims

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return claims, errInvalidToken
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return claims, errInvalidToken
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return claims, errInvalidToken
	}

	return claims, nil
}

func (claims keycloakClaims) roles() map[string]struct{} {
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

func (claims TokenClaims) hasRole(role string) bool {
	_, ok := claims.Roles[role]
	return ok
}

func (claims TokenClaims) hasAudience(audience string) bool {
	for _, value := range claims.Audience {
		if value == audience {
			return true
		}
	}
	return false
}

func (audience audienceClaim) has(value string) bool {
	for _, audience := range audience {
		if audience == value {
			return true
		}
	}
	return false
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
	case errors.Is(err, errInvalidSubject):
		return "invalid_subject"
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
