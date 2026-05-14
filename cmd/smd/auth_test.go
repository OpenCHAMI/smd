// MIT License
//
// (C) Copyright [2026] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	jwtauth "github.com/OpenCHAMI/jwtauth/v5"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/openchami/tokensmith/pkg/authn"
	"github.com/openchami/tokensmith/pkg/authz"
	"github.com/openchami/tokensmith/pkg/keys"
)

func TestInitializeAuthClearsStaleStateOnError(t *testing.T) {
	s := &SmD{
		jwksURL:                  "https://example.com/jwks.json",
		authBackend:              authBackendTokenSmith,
		protectedAuthMiddleware:  func(next http.Handler) http.Handler { return next },
		protectedAuthzMiddleware: func(next http.Handler) http.Handler { return next },
		legacyTokenAuth:          jwtauth.New("HS256", []byte("secret"), nil),
	}

	err := s.initializeAuth()
	if err == nil {
		t.Fatal("expected initializeAuth to fail without TokenSmith issuer and audiences")
	}
	if s.protectedAuthMiddleware != nil {
		t.Fatal("expected initializeAuth to clear stale protected auth middleware on failure")
	}
	if s.legacyTokenAuth != nil {
		t.Fatal("expected initializeAuth to clear stale legacy auth state on failure")
	}
	if s.protectedAuthzMiddleware != nil {
		t.Fatal("expected initializeAuth to clear stale protected authz middleware on failure")
	}
}

func TestProtectedAuthMiddlewareFallsClosed(t *testing.T) {
	s := &SmD{jwksURL: "https://example.com/jwks.json"}
	middleware := s.ProtectedAuthMiddleware()
	if middleware == nil {
		t.Fatal("expected protected auth middleware when authentication is enabled")
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recorder := httptest.NewRecorder()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
}

func TestBuildLegacyProtectedAuthMiddlewareRejectsMissingToken(t *testing.T) {
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	middleware := buildLegacyProtectedAuthMiddleware(tokenAuth)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recorder := httptest.NewRecorder()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestExtractScopesFromClaims(t *testing.T) {
	claims := map[string]any{
		"scp":   []any{"alpha", 42, "beta"},
		"scope": "gamma delta",
	}

	scopes := extractScopesFromClaims(claims)
	expected := []string{"alpha", "beta", "gamma", "delta"}
	if len(scopes) != len(expected) {
		t.Fatalf("expected %d scopes, got %d: %v", len(expected), len(scopes), scopes)
	}
	for index, scope := range expected {
		if scopes[index] != scope {
			t.Fatalf("expected scope %q at index %d, got %q", scope, index, scopes[index])
		}
	}
}

func TestVerifyClaimsUsesLegacyContext(t *testing.T) {
	s := &SmD{}
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	token := jwt.New()
	if err := token.Set("sub", "node-1"); err != nil {
		t.Fatalf("failed to set sub claim: %v", err)
	}
	if err := token.Set("iss", "issuer"); err != nil {
		t.Fatalf("failed to set iss claim: %v", err)
	}
	req = req.WithContext(context.WithValue(req.Context(), jwtauth.TokenCtxKey, token))

	ok, err := s.VerifyClaims([]string{"sub", "iss"}, req)
	if err != nil {
		t.Fatalf("expected VerifyClaims to succeed, got error: %v", err)
	}
	if !ok {
		t.Fatal("expected VerifyClaims to return true")
	}
}

func TestVerifyScopeSupportsMixedScopeClaims(t *testing.T) {
	s := &SmD{}
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	token := jwt.New()
	if err := token.Set("scp", []string{"alpha", "beta"}); err != nil {
		t.Fatalf("failed to set scp claim: %v", err)
	}
	if err := token.Set("scope", "gamma delta"); err != nil {
		t.Fatalf("failed to set scope claim: %v", err)
	}
	req = req.WithContext(context.WithValue(req.Context(), jwtauth.TokenCtxKey, token))

	ok, err := s.VerifyScope([]string{"alpha", "delta"}, req)
	if err != nil {
		t.Fatalf("expected VerifyScope to succeed, got error: %v", err)
	}
	if !ok {
		t.Fatal("expected VerifyScope to return true")
	}
}

func TestTokenSmithAuthRejectionLoggingIncludesClearContext(t *testing.T) {
	var logBuffer bytes.Buffer
	s := &SmD{
		lg:            log.New(&logBuffer, "", 0),
		authIssuer:    "https://issuer.example.com",
		authAudiences: []string{"smd", "admin"},
	}
	middleware := s.withAuthRejectionLogging(authBackendTokenSmith, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			http.Error(w, "issuer mismatch", http.StatusUnauthorized)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.RemoteAddr = "10.2.3.4:12345"
	req.Header.Set("Authorization", "Bearer redacted-token")
	recorder := httptest.NewRecorder()

	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, req)

	output := logBuffer.String()
	checks := []string{
		"Auth rejection backend=tokensmith status=401",
		"method=GET",
		"path=/protected",
		"remote=10.2.3.4:12345",
		"auth_header=present",
		"auth_scheme=bearer",
		"auth_constraints=configured",
		"issuer_configured=true",
		"audiences_configured=true",
		"challenge=present",
		"challenge_scheme=bearer",
		"challenge_error=invalid_token",
		"auth_reason=issuer",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected log output to contain %q, got %q", check, output)
		}
	}
	if strings.Contains(output, "redacted-token") {
		t.Fatalf("expected auth rejection log not to contain bearer token, got %q", output)
	}
	blocked := []string{
		"https://issuer.example.com",
		"smd,admin",
		`www_authenticate="Bearer error=\"invalid_token\""`,
		`detail="issuer mismatch"`,
		"issuer mismatch",
	}
	for _, blockedValue := range blocked {
		if strings.Contains(output, blockedValue) {
			t.Fatalf("expected auth rejection log not to contain %q, got %q", blockedValue, output)
		}
	}
}

func TestTokenSmithAuthRejectionLoggingSkipsSuccessfulRequests(t *testing.T) {
	var logBuffer bytes.Buffer
	s := &SmD{lg: log.New(&logBuffer, "", 0)}
	middleware := s.withAuthRejectionLogging(authBackendTokenSmith, func(next http.Handler) http.Handler {
		return next
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recorder := httptest.NewRecorder()
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, req)

	if output := logBuffer.String(); output != "" {
		t.Fatalf("expected no auth rejection log for successful request, got %q", output)
	}
}

func TestTokenSmithJWKSBackedRejectedTokenLogsClearly(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	kid := rsaThumbprint(t, &privateKey.PublicKey)
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(smdTestJWKSJSON(t, kid, &privateKey.PublicKey))
	}))
	defer jwksServer.Close()

	var logBuffer bytes.Buffer
	s := &SmD{
		lg:            log.New(&logBuffer, "", 0),
		jwksURL:       jwksServer.URL,
		authBackend:   authBackendTokenSmith,
		authIssuer:    "https://issuer.example.com",
		authAudiences: []string{"smd"},
	}

	if err := s.initializeTokenSmithAuth(); err != nil {
		t.Fatalf("failed to initialize TokenSmith auth: %v", err)
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, validTokenSmithServiceClaims("https://wrong-issuer.example.com", []string{"smd"}, "node-1"))
	token.Header["kid"] = kid
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign JWT: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.RemoteAddr = "10.20.30.40:12345"
	req.Header.Set("Authorization", "Bearer "+signedToken)
	recorder := httptest.NewRecorder()

	handler := s.ProtectedAuthMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusUnauthorized, recorder.Code, recorder.Body.String())
	}

	output := logBuffer.String()
	checks := []string{
		"Auth rejection backend=tokensmith status=401",
		"method=GET",
		"path=/protected",
		"remote=10.20.30.40:12345",
		"auth_header=present",
		"auth_scheme=bearer",
		"auth_constraints=configured",
		"issuer_configured=true",
		"audiences_configured=true",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected log output to contain %q, got %q", check, output)
		}
	}
	if strings.Contains(output, signedToken) {
		t.Fatalf("expected log output not to include bearer token, got %q", output)
	}
	blocked := []string{
		s.authIssuer,
		strings.Join(s.authAudiences, ","),
	}
	for _, blockedValue := range blocked {
		if blockedValue != "" && strings.Contains(output, blockedValue) {
			t.Fatalf("expected log output not to include %q, got %q", blockedValue, output)
		}
	}
}

func TestTokenSmithJWKSBackedAcceptedTokenSetsClaimsAndSkipsRejectionLog(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	kid := rsaThumbprint(t, &privateKey.PublicKey)
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(smdTestJWKSJSON(t, kid, &privateKey.PublicKey))
	}))
	defer jwksServer.Close()

	var logBuffer bytes.Buffer
	s := &SmD{
		lg:            log.New(&logBuffer, "", 0),
		jwksURL:       jwksServer.URL,
		authBackend:   authBackendTokenSmith,
		authIssuer:    "https://issuer.example.com",
		authAudiences: []string{"smd"},
	}

	if err := s.initializeTokenSmithAuth(); err != nil {
		t.Fatalf("failed to initialize TokenSmith auth: %v", err)
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, validTokenSmithServiceClaims("https://issuer.example.com", []string{"smd"}, "node-1"))
	token.Header["kid"] = kid
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign JWT: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.RemoteAddr = "10.20.30.41:12345"
	req.Header.Set("Authorization", "Bearer "+signedToken)
	recorder := httptest.NewRecorder()

	handlerCalled := false
	handler := s.ProtectedAuthMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		claims, ok := authn.VerifiedClaimsFromContext(r.Context())
		if !ok {
			t.Fatal("expected verified claims in TokenSmith request context")
		}
		if claims["sub"] != "node-1" {
			t.Fatalf("expected subject claim node-1, got %#v", claims["sub"])
		}
		principal, ok := authz.PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("expected TokenSmith principal in request context")
		}
		if principal.ID != "node-1" {
			t.Fatalf("expected principal ID node-1, got %q", principal.ID)
		}
		if !containsString(principal.Roles, "service") {
			t.Fatalf("expected TokenSmith principal roles to include service, got %v", principal.Roles)
		}

		verified, err := s.VerifyClaims([]string{"sub", "iss", "aud"}, r)
		if err != nil {
			t.Fatalf("expected VerifyClaims to succeed, got error: %v", err)
		}
		if !verified {
			t.Fatal("expected VerifyClaims to return true")
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	handler.ServeHTTP(recorder, req)

	if !handlerCalled {
		t.Fatal("expected protected handler to be called for valid TokenSmith JWT")
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusNoContent, recorder.Code, recorder.Body.String())
	}
	if output := logBuffer.String(); output != "" {
		t.Fatalf("expected no auth rejection log for valid TokenSmith JWT, got %q", output)
	}
}

func smdTestJWKSJSON(t *testing.T, kid string, pub *rsa.PublicKey) []byte {
	t.Helper()

	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())

	obj := map[string]any{
		"keys": []any{
			map[string]any{
				"kty": "RSA",
				"kid": kid,
				"alg": "RS256",
				"use": "sig",
				"n":   n,
				"e":   e,
			},
		},
	}
	b, err := json.Marshal(obj)
	if err != nil {
		t.Fatalf("failed to marshal JWKS: %v", err)
	}
	return b
}

func TestNewRouterUsesProtectedAuthMiddleware(t *testing.T) {
	called := false
	s := &SmD{
		jwksURL: "https://example.com/jwks.json",
		protectedAuthMiddleware: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				next.ServeHTTP(w, r)
			})
		},
	}

	router := s.NewRouter(nil, Routes{{
		Name:    "protected",
		Method:  http.MethodGet,
		Pattern: "/protected",
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	}})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if !called {
		t.Fatal("expected protected auth middleware to wrap protected routes")
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func TestProtectedAuthzMiddlewareFallsClosed(t *testing.T) {
	s := &SmD{jwksURL: "https://example.com/jwks.json", authzMode: "enforce"}
	middleware := s.ProtectedAuthzMiddleware()
	if middleware == nil {
		t.Fatal("expected protected authz middleware when authorization is enabled")
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recorder := httptest.NewRecorder()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}
}

func TestLegacyProtectedAuthMiddlewareSetsPrincipal(t *testing.T) {
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	middleware := withLegacyPrincipal(buildLegacyProtectedAuthMiddleware(tokenAuth))
	_, signedToken, err := tokenAuth.Encode(map[string]any{
		"sub":   "legacy-user",
		"iss":   "issuer",
		"aud":   "smd",
		"roles": []string{"viewer"},
	})
	if err != nil {
		t.Fatalf("failed to sign legacy token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken)
	recorder := httptest.NewRecorder()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := authz.PrincipalFromContext(r.Context())
		if !ok {
			t.Fatal("expected legacy principal in context")
		}
		if principal.ID != "legacy-user" {
			t.Fatalf("expected principal ID legacy-user, got %q", principal.ID)
		}
		if !containsString(principal.Roles, "viewer") {
			t.Fatalf("expected principal roles to include viewer, got %v", principal.Roles)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d with body %q", http.StatusNoContent, recorder.Code, recorder.Body.String())
	}
}

func TestTokenSmithAuthzEnforceAllowsServiceTokenAndDeniesViewerWrite(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	kid := rsaThumbprint(t, &privateKey.PublicKey)
	jwksServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(smdTestJWKSJSON(t, kid, &privateKey.PublicKey))
	}))
	defer jwksServer.Close()

	policyDir := writeAuthzPolicyDir(t, strings.Join([]string{
		"p, role:viewer, /*, read",
		"p, role:service, /*, read",
		"p, role:service, /*, write",
	}, "\n")+"\n")

	s := &SmD{
		jwksURL:        jwksServer.URL,
		authBackend:    authBackendTokenSmith,
		authIssuer:     "https://issuer.example.com",
		authAudiences:  []string{"smd"},
		authzMode:      "enforce",
		authzPolicyDir: policyDir,
	}

	if err := s.initializeAuth(); err != nil {
		t.Fatalf("failed to initialize auth stack: %v", err)
	}

	serviceToken := signToken(t, privateKey, kid, validTokenSmithServiceClaims("https://issuer.example.com", []string{"smd"}, "svc-a"))
	viewerToken := signToken(t, privateKey, kid, validTokenSmithRoleClaims("https://issuer.example.com", []string{"smd"}, "user-a", []string{"viewer"}))

	protectedHandler := s.ProtectedAuthMiddleware()(s.ProtectedAuthzMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})))

	allowReq := httptest.NewRequest(http.MethodPost, "/hsm/v2/State/Components", nil)
	allowReq.Header.Set("Authorization", "Bearer "+serviceToken)
	allowRecorder := httptest.NewRecorder()
	protectedHandler.ServeHTTP(allowRecorder, allowReq)
	if allowRecorder.Code != http.StatusNoContent {
		t.Fatalf("expected service token write to be allowed, got %d with body %q", allowRecorder.Code, allowRecorder.Body.String())
	}

	denyReq := httptest.NewRequest(http.MethodPost, "/hsm/v2/State/Components", nil)
	denyReq.Header.Set("Authorization", "Bearer "+viewerToken)
	denyRecorder := httptest.NewRecorder()
	protectedHandler.ServeHTTP(denyRecorder, denyReq)
	if denyRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected viewer write to be denied, got %d with body %q", denyRecorder.Code, denyRecorder.Body.String())
	}
	if !strings.Contains(denyRecorder.Body.String(), "AUTHZ_DENIED") {
		t.Fatalf("expected authz deny response, got %q", denyRecorder.Body.String())
	}
}

func TestNewRouterUsesProtectedAuthAndAuthzMiddleware(t *testing.T) {
	authCalled := false
	authzCalled := false
	s := &SmD{
		jwksURL:   "https://example.com/jwks.json",
		authzMode: "enforce",
		protectedAuthMiddleware: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				authCalled = true
				next.ServeHTTP(w, r)
			})
		},
		protectedAuthzMiddleware: func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				authzCalled = true
				next.ServeHTTP(w, r)
			})
		},
	}

	router := s.NewRouter(nil, Routes{{
		Name:    "protected",
		Method:  http.MethodGet,
		Pattern: "/protected",
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	}})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if !authCalled {
		t.Fatal("expected protected auth middleware to wrap protected routes")
	}
	if !authzCalled {
		t.Fatal("expected protected authz middleware to wrap protected routes")
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}
}

func validTokenSmithServiceClaims(issuer string, audience []string, subject string) jwtv5.MapClaims {
	now := time.Now().UTC()
	iat := now.Add(-time.Minute)
	return jwtv5.MapClaims{
		"sub":          subject,
		"iss":          issuer,
		"aud":          audience,
		"iat":          iat.Unix(),
		"nbf":          iat.Unix(),
		"exp":          now.Add(time.Minute).Unix(),
		"auth_level":   "IAL2",
		"auth_factors": 2,
		"auth_methods": []string{"service", "certificate"},
		"session_id":   "service-" + subject,
		"session_exp":  iat.Add(24 * time.Hour).Unix(),
		"auth_events":  []string{"service_auth"},
	}
}

func validTokenSmithRoleClaims(issuer string, audience []string, subject string, roles []string) jwtv5.MapClaims {
	claims := validTokenSmithServiceClaims(issuer, audience, subject)
	claims["auth_events"] = []string{"login"}
	claims["auth_methods"] = []string{"password", "mfa"}
	claims["roles"] = roles
	return claims
}

func signToken(t *testing.T, privateKey *rsa.PrivateKey, kid string, claims jwtv5.MapClaims) string {
	t.Helper()
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("failed to sign JWT: %v", err)
	}
	return signedToken
}

func writeAuthzPolicyDir(t *testing.T, policyCSV string) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"model.conf": strings.Join([]string{
			"[request_definition]",
			"r = sub, obj, act",
			"",
			"[policy_definition]",
			"p = sub, obj, act",
			"",
			"[role_definition]",
			"g = _, _",
			"",
			"[policy_effect]",
			"e = some(where (p.eft == allow))",
			"",
			"[matchers]",
			"m = g(r.sub, p.sub) && keyMatch2(r.obj, p.obj) && r.act == p.act",
		}, "\n") + "\n",
		"policy.csv":   policyCSV,
		"grouping.csv": "# no role inheritance in test policy\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}
	return dir
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func rsaThumbprint(t *testing.T, pub *rsa.PublicKey) string {
	t.Helper()
	kid, err := keys.RFC7638Thumbprint(pub)
	if err != nil {
		t.Fatalf("failed to compute RFC7638 thumbprint: %v", err)
	}
	return kid
}
