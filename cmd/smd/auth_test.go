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
	"strings"
	"testing"
	"time"

	jwtauth "github.com/OpenCHAMI/jwtauth/v5"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/openchami/tokensmith/pkg/authn"
)

func TestInitializeAuthClearsStaleStateOnError(t *testing.T) {
	s := &SmD{
		jwksURL:                 "https://example.com/jwks.json",
		authBackend:             authBackendTokenSmith,
		protectedAuthMiddleware: func(next http.Handler) http.Handler { return next },
		legacyTokenAuth:         jwtauth.New("HS256", []byte("secret"), nil),
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

	kid := "test-key-1"
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

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, jwtv5.MapClaims{
		"sub": "node-1",
		"iss": "https://wrong-issuer.example.com",
		"aud": []string{"smd"},
		"iat": time.Now().Add(-time.Minute).Unix(),
		"nbf": time.Now().Add(-time.Minute).Unix(),
		"exp": time.Now().Add(time.Minute).Unix(),
	})
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

	kid := "test-key-2"
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

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, jwtv5.MapClaims{
		"sub": "node-1",
		"iss": "https://issuer.example.com",
		"aud": []string{"smd"},
		"iat": time.Now().Add(-time.Minute).Unix(),
		"nbf": time.Now().Add(-time.Minute).Unix(),
		"exp": time.Now().Add(time.Minute).Unix(),
	})
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
