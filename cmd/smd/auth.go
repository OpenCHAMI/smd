package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"

	jwtauth "github.com/OpenCHAMI/jwtauth/v5"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
	openchami_authenticator "github.com/openchami/chi-middleware/auth"
	"github.com/openchami/tokensmith/pkg/authn"
	"github.com/openchami/tokensmith/pkg/authz"
	"github.com/openchami/tokensmith/pkg/authz/engine"
)

var requiredLegacyClaims = []string{"sub", "iss", "aud"}

func (s *SmD) IsUsingAuthentication() bool {
	return s.jwksURL != ""
}

func (s *SmD) UsingTokenSmithAuth() bool {
	return s.IsUsingAuthentication() && s.authBackend == authBackendTokenSmith
}

func (s *SmD) IsUsingAuthorization() bool {
	if !s.IsUsingAuthentication() {
		return false
	}
	mode, err := parseAuthzMode(s.authzMode)
	return err == nil && mode != authz.ModeOff
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

	var err error
	if s.UsingTokenSmithAuth() {
		err = s.initializeTokenSmithAuth()
	} else {
		err = s.initializeLegacyAuth()
	}
	if err != nil {
		return err
	}
	if s.IsUsingAuthorization() {
		if err := s.initializeAuthz(); err != nil {
			s.clearAuthState()
			return err
		}
	}

	return nil
}

func (s *SmD) clearAuthState() {
	s.legacyTokenAuth = nil
	s.protectedAuthMiddleware = nil
	s.protectedAuthzMiddleware = nil
}

func (s *SmD) initializeLegacyAuth() error {
	tokenAuth, err := fetchLegacyTokenAuthFromURL(s.jwksURL)
	if err != nil {
		return err
	}

	s.legacyTokenAuth = tokenAuth
	s.protectedAuthMiddleware = s.withAuthRejectionLogging(authBackendLegacy, withLegacyPrincipal(buildLegacyProtectedAuthMiddleware(tokenAuth)))
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
		Mapper: func(ctx context.Context, token *jwtv5.Token, claims jwtv5.MapClaims) (authz.Principal, error) {
			return principalFromClaims(map[string]any(claims)), nil
		},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize TokenSmith middleware: %w", err)
	}

	s.protectedAuthMiddleware = s.withAuthRejectionLogging(authBackendTokenSmith, mw)
	return nil
}

func (s *SmD) initializeAuthz() error {
	mode, err := parseAuthzMode(s.authzMode)
	if err != nil {
		return err
	}
	if mode == authz.ModeOff {
		return nil
	}

	policyDir := strings.TrimSpace(s.authzPolicyDir)
	if policyDir == "" {
		return errors.New("authz-policy-dir is required when authz-mode is enabled")
	}

	authorizer, err := engine.NewBuilder().
		WithModelPath(filepath.Join(policyDir, "model.conf")).
		WithPolicyPath(filepath.Join(policyDir, "policy.csv")).
		WithGroupingPath(filepath.Join(policyDir, "grouping.csv")).
		Build()
	if err != nil {
		return fmt.Errorf("failed to initialize TokenSmith authorization: %w", err)
	}

	s.protectedAuthzMiddleware = authz.NewMiddleware(
		authorizer,
		authz.PathMethodMapper{MethodToAction: authz.MethodToActionREST()},
		authz.WithMode(mode),
		authz.WithRequireAuthn(true),
	).Handler

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

func (s *SmD) ProtectedAuthzMiddleware() func(http.Handler) http.Handler {
	if !s.IsUsingAuthorization() {
		return nil
	}
	if s.protectedAuthzMiddleware != nil {
		return s.protectedAuthzMiddleware
	}

	return unavailableAuthzMiddleware()
}

func unavailableAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sendJsonError(w, http.StatusServiceUnavailable, "authentication is enabled but not initialized")
		})
	}
}

func unavailableAuthzMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sendJsonError(w, http.StatusServiceUnavailable, "authorization is enabled but not initialized")
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

func withLegacyPrincipal(base func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return base(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := legacyVerifiedClaims(r.Context())
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			principal := principalFromClaims(claims)
			if strings.TrimSpace(principal.ID) == "" {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := authz.SetPrincipal(r.Context(), &principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))
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
		message += fmt.Sprintf(
			" auth_constraints=configured issuer_configured=%t audiences_configured=%t",
			strings.TrimSpace(s.authIssuer) != "",
			len(s.authAudiences) > 0,
		)
	}
	if challengeState, challengeScheme, challengeError := summarizeWWWAuthenticate(recorder.Header().Get("WWW-Authenticate")); challengeState != "missing" {
		message += fmt.Sprintf(" challenge=%s challenge_scheme=%s", challengeState, challengeScheme)
		if challengeError != "none" {
			message += fmt.Sprintf(" challenge_error=%s", challengeError)
		}
	}
	if detailClass := classifyAuthFailureDetail(recorder.BodySnippet(), statusCode); detailClass != "none" {
		message += fmt.Sprintf(" auth_reason=%s", detailClass)
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

func summarizeWWWAuthenticate(header string) (state string, scheme string, errorCode string) {
	challenge := compactWhitespace(header)
	if challenge == "" {
		return "missing", "none", "none"
	}

	fields := strings.Fields(challenge)
	if len(fields) == 0 {
		return "present", "unknown", "none"
	}

	scheme = strings.ToLower(fields[0])
	errorCode = "none"
	if index := strings.Index(strings.ToLower(challenge), `error="`); index >= 0 {
		value := challenge[index+len(`error="`):]
		if end := strings.Index(value, `"`); end >= 0 {
			parsed := strings.TrimSpace(value[:end])
			if parsed != "" {
				errorCode = strings.ToLower(parsed)
			}
		}
	}

	return "present", scheme, errorCode
}

func classifyAuthFailureDetail(detail string, statusCode int) string {
	if statusCode == http.StatusForbidden {
		return "forbidden"
	}

	normalized := strings.ToLower(compactWhitespace(detail))
	if normalized == "" {
		return "none"
	}

	switch {
	case strings.Contains(normalized, "issuer"):
		return "issuer"
	case strings.Contains(normalized, "audience") || strings.Contains(normalized, " aud"):
		return "audience"
	case strings.Contains(normalized, "scope"):
		return "scope"
	case strings.Contains(normalized, "claim"):
		return "claims"
	case strings.Contains(normalized, "expir"):
		return "expired"
	case strings.Contains(normalized, "token"):
		return "token"
	default:
		return "rejected"
	}
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

	claims, err := legacyVerifiedClaims(r.Context())
	if err != nil {
		return nil, err
	}
	return claims, nil
}

func legacyVerifiedClaims(ctx context.Context) (map[string]any, error) {
	_, claims, err := jwtauth.FromContext(ctx)
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

func extractStringClaimValues(claims map[string]any, keys ...string) []string {
	var values []string
	appendValues := func(raw any) {
		switch typed := raw.(type) {
		case []any:
			for _, value := range typed {
				if asString, ok := value.(string); ok && strings.TrimSpace(asString) != "" {
					values = append(values, strings.TrimSpace(asString))
				}
			}
		case []string:
			for _, value := range typed {
				if strings.TrimSpace(value) != "" {
					values = append(values, strings.TrimSpace(value))
				}
			}
		case string:
			for _, value := range strings.Fields(typed) {
				if strings.TrimSpace(value) != "" {
					values = append(values, strings.TrimSpace(value))
				}
			}
		}
	}

	for _, key := range keys {
		if raw, ok := claims[key]; ok {
			appendValues(raw)
		}
	}

	return values
}

func principalFromClaims(claims map[string]any) authz.Principal {
	principal := authz.Principal{}
	if sub, ok := claims["sub"].(string); ok {
		principal.ID = strings.TrimSpace(sub)
	}
	principal.Roles = append(principal.Roles, extractStringClaimValues(claims, "roles", "role")...)
	if hasServiceAuthEvent(claims) {
		principal.Roles = append(principal.Roles, "service")
	}
	principal.Roles = dedupeStrings(principal.Roles)
	return principal
}

func hasServiceAuthEvent(claims map[string]any) bool {
	for _, event := range extractStringClaimValues(claims, "auth_events") {
		if strings.EqualFold(event, "service_auth") {
			return true
		}
	}
	return false
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func parseAuthzMode(raw string) (authz.Mode, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "off":
		return authz.ModeOff, nil
	case "shadow":
		return authz.ModeShadow, nil
	case "enforce":
		return authz.ModeEnforce, nil
	default:
		return authz.ModeOff, fmt.Errorf("unsupported authz mode %q", raw)
	}
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
