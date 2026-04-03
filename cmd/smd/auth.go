package main

import (
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
	"github.com/openchami/tokensmith/pkg/authn"
)

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
	if s.UsingTokenSmithAuth() {
		return s.initializeTokenSmithAuth()
	}

	return s.fetchPublicKeyFromURL(s.jwksURL)
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

	s.authMiddleware = mw
	return nil
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
	var scopes []string

	appendScopes := func(slice []string, scopeClaim any) []string {
		switch scopeClaim.(type) {
		case []any:
			// convert all scopes to str and append
			for _, s := range scopeClaim.([]any) {
				switch s.(type) {
				case string:
					slice = append(slice, s.(string))
				}
			}
		case []string:
			slice = append(slice, scopeClaim.([]string)...)
		case string:
			slice = append(slice, strings.Fields(scopeClaim.(string))...)
		}
		return slice
	}
	if v, ok := claims["scp"]; ok {
		scopes = appendScopes(scopes, v)
	}
	if v, ok := claims["scope"]; ok {
		scopes = appendScopes(scopes, v)
	}

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

func (s *SmD) fetchPublicKeyFromURL(url string) error {
	client := newHTTPClient()

	jwks, err := fetchJWKSFromURL(url, client)
	if err != nil {
		return err
	}
	s.tokenAuth, err = jwtauth.NewKeySet(jwks)
	if err != nil {
		return fmt.Errorf("failed to initialize JWKS: %v", err)
	}

	return nil
}
