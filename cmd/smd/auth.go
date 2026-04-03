package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	jwtauth "github.com/OpenCHAMI/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
	openchami_authenticator "github.com/openchami/chi-middleware/auth"
	"github.com/openchami/tokensmith/pkg/authn"
)

var requiredLegacyClaims = []string{"sub", "iss", "aud"}

func (s *SmD) IsUsingAuthentication() bool {
	return s.jwksURL != ""
}

func (s *SmD) UsingTokenSmithAuth() bool {
	return s.IsUsingAuthentication() && s.authBackend == authBackendTokenSmith
}

func parseCSVValues(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	values := strings.Split(raw, ",")
	parsed := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		parsed = append(parsed, value)
	}

	return parsed
}

func (s *SmD) initializeAuth() error {
	s.clearAuthState()

	if s.UsingTokenSmithAuth() {
		return s.initializeTokenSmithAuth()
	}

	return s.initializeLegacyAuth()
}

func (s *SmD) clearAuthState() {
	s.legacyTokenAuth = nil
	s.protectedAuthMiddleware = nil
}

func (s *SmD) initializeLegacyAuth() error {
	tokenAuth, err := fetchLegacyTokenAuthFromURL(s.jwksURL)
	if err != nil {
		return err
	}

	s.legacyTokenAuth = tokenAuth
	s.protectedAuthMiddleware = s.withAuthRejectionLogging(authBackendLegacy, buildLegacyProtectedAuthMiddleware(tokenAuth))
	return nil
}

func (s *SmD) initializeTokenSmithAuth() error {
	if strings.TrimSpace(s.authIssuer) == "" {
		return errors.New("auth-issuer is required when auth-backend=tokensmith")
	}
	if len(s.authAudiences) == 0 {
		return errors.New("auth-audiences is required when auth-backend=tokensmith")
	}

	client := newHTTPClient()
	client.Timeout = 40 * time.Second
	if _, err := fetchJWKSFromURL(s.jwksURL, client); err != nil {
		return err
	}

	mw, err := authn.Middleware(authn.Options{
		Issuers:    []string{s.authIssuer},
		Audiences:  append([]string(nil), s.authAudiences...),
		JWKSURLs:   []string{s.jwksURL},
		HTTPClient: client,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize TokenSmith middleware: %w", err)
	}

	s.protectedAuthMiddleware = s.withAuthRejectionLogging(authBackendTokenSmith, mw)
	return nil
}

func (s *SmD) ProtectedAuthMiddleware() func(http.Handler) http.Handler {
	if !s.IsUsingAuthentication() {
		return nil
	}
	if s.protectedAuthMiddleware != nil {
		return s.protectedAuthMiddleware
	}

	return unavailableAuthMiddleware()
}

func unavailableAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sendJsonError(w, http.StatusServiceUnavailable, "authentication is enabled but not initialized")
		})
	}
}

func buildLegacyProtectedAuthMiddleware(tokenAuth *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return jwtauth.Verifier(tokenAuth)(
			openchami_authenticator.AuthenticatorWithRequiredClaims(tokenAuth, requiredLegacyClaims)(next),
		)
	}
}

func (s *SmD) withAuthRejectionLogging(backend string, base func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := base(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := &authStatusRecorder{ResponseWriter: w}
			handler.ServeHTTP(recorder, r)
			s.logAuthRejection(backend, r, recorder)
		})
	}
}

func (s *SmD) logAuthRejection(backend string, r *http.Request, recorder *authStatusRecorder) {
	statusCode := recorder.StatusCode()
	if statusCode != http.StatusUnauthorized && statusCode != http.StatusForbidden {
		return
	}
	if s == nil || s.lg == nil {
		return
	}

	authHeaderState, authHeaderScheme := describeAuthorizationHeader(r.Header.Get("Authorization"))
	message := fmt.Sprintf(
		"Auth rejection backend=%s status=%d method=%s path=%s remote=%s auth_header=%s auth_scheme=%s",
		backend,
		statusCode,
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		authHeaderState,
		authHeaderScheme,
	)

	if backend == authBackendTokenSmith {
		message += fmt.Sprintf(" expected_issuer=%q expected_audiences=%q", s.authIssuer, strings.Join(s.authAudiences, ","))
	}
	if challenge := compactWhitespace(recorder.Header().Get("WWW-Authenticate")); challenge != "" {
		message += fmt.Sprintf(" www_authenticate=%q", challenge)
	}
	if detail := recorder.BodySnippet(); detail != "" {
		message += fmt.Sprintf(" detail=%q", detail)
	}

	s.LogAlways("%s", message)
}

func describeAuthorizationHeader(header string) (state string, scheme string) {
	trimmed := strings.TrimSpace(header)
	if trimmed == "" {
		return "missing", "none"
	}

	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "present", "unknown"
	}

	return "present", strings.ToLower(fields[0])
}

func compactWhitespace(value string) string {
	return strings.Join(strings.Fields(value), " ")
}

type authStatusRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (w *authStatusRecorder) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *authStatusRecorder) Write(body []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	remaining := 256 - w.body.Len()
	if remaining > 0 {
		if len(body) > remaining {
			_, _ = w.body.Write(body[:remaining])
		} else {
			_, _ = w.body.Write(body)
		}
	}

	return w.ResponseWriter.Write(body)
}

func (w *authStatusRecorder) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}

	return w.statusCode
}

func (w *authStatusRecorder) BodySnippet() string {
	return compactWhitespace(w.body.String())
}

func (s *SmD) verifiedClaimsFromRequest(r *http.Request) (map[string]any, error) {
	if s.UsingTokenSmithAuth() {
		claims, ok := authn.VerifiedClaimsFromContext(r.Context())
		if !ok {
			return nil, errors.New("verified claims not found in context")
		}
		return claims, nil
	}

	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func (s *SmD) VerifyClaims(testClaims []string, r *http.Request) (bool, error) {
	claims, err := s.verifiedClaimsFromRequest(r)
	if err != nil {
		return false, fmt.Errorf("failed to get claims(s) from token: %v", err)
	}

	// verify that each one of the test claims are included
	for _, testClaim := range testClaims {
		_, ok := claims[testClaim]
		if !ok {
			return false, fmt.Errorf("failed to verify claim(s) from token: %s", testClaim)
		}
	}
	return true, nil
}

func (s *SmD) VerifyScope(testScopes []string, r *http.Request) (bool, error) {
	claims, err := s.verifiedClaimsFromRequest(r)
	if err != nil {
		return false, fmt.Errorf("failed to get claim(s) from token: %v", err)
	}
	scopes := extractScopesFromClaims(claims)

	// verify that each of the test scopes are included
	for _, testScope := range testScopes {
		index := slices.Index(scopes, testScope)
		if index < 0 {
			return false, fmt.Errorf("invalid or missing scope")
		}
	}
	// NOTE: should this be ok if no scopes were found?
	return true, nil
}

func extractScopesFromClaims(claims map[string]any) []string {
	appendScopes := func(slice []string, scopeClaim any) []string {
		switch typedScopeClaim := scopeClaim.(type) {
		case []any:
			for _, scope := range typedScopeClaim {
				if typedScope, ok := scope.(string); ok {
					slice = append(slice, typedScope)
				}
			}
		case []string:
			slice = append(slice, typedScopeClaim...)
		case string:
			slice = append(slice, strings.Fields(typedScopeClaim)...)
		}
		return slice
	}

	var scopes []string
	if v, ok := claims["scp"]; ok {
		scopes = appendScopes(scopes, v)
	}
	if v, ok := claims["scope"]; ok {
		scopes = appendScopes(scopes, v)
	}

	return scopes
}

type statusCheckTransport struct {
	http.RoundTripper
}

func (ct *statusCheckTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err == nil && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return resp, err
}

func newHTTPClient() *http.Client {
	return &http.Client{Transport: &statusCheckTransport{}}
}

func fetchJWKSFromURL(url string, client *http.Client) ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	set, err := jwk.Fetch(ctx, url, jwk.WithHTTPClient(client))
	if err != nil {
		msg := "%w"

		// if the error tree contains an EOF, it means that the response was empty,
		// so add a more descriptive message to the error tree
		if errors.Is(err, io.EOF) {
			msg = "received empty response for key: %w"
		}

		return nil, fmt.Errorf(msg, err)
	}
	jwks, err := json.Marshal(set)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JWKS: %v", err)
	}

	return jwks, nil
}

func fetchLegacyTokenAuthFromURL(url string) (*jwtauth.JWTAuth, error) {
	client := newHTTPClient()

	jwks, err := fetchJWKSFromURL(url, client)
	if err != nil {
		return nil, err
	}
	tokenAuth, err := jwtauth.NewKeySet(jwks)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWKS: %v", err)
	}

	return tokenAuth, nil
}
