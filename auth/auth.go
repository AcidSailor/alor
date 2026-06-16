// Package auth provides an oauth2.TokenSource for Alor's non-RFC /refresh
// endpoint. Construct a [TokenSource] via [New] and pass it to alor.NewClient.
package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/acidsailor/restkit"
	"golang.org/x/oauth2"
)

// Alor OAuth hosts — a dedicated subdomain, not the alor.Endpoint* API hosts.
const (
	HostProduction = "https://oauth.alor.ru"
	HostTest       = "https://oauthdev.alor.ru"
)

const refreshPath = "/refresh"

// clientName tags every restkit error's Name field so /refresh failures are
// attributable to auth, distinct from the alor API client.
const clientName = "alor-auth"

// New returns an auto-refreshing [oauth2.TokenSource] that mints Alor JWT access
// tokens by POSTing refreshToken to Alor's custom /refresh endpoint. Alor's
// refresh is not RFC 6749 §6 compliant, so x/oauth2's built-in refresh cannot
// drive it; this source fills the gap while staying TokenSource-compatible.
//
// oauthHost selects the OAuth host: [HostProduction] or [HostTest]. refreshToken
// is either the long-lived token from https://alor.dev/myapps (JWT flow) or the
// refresh_token from oauth2.Config.Exchange (OAuth code flow); both refresh
// identically.
//
// The /refresh exchange runs over auth's own restkit transport, separate from
// the alor API client's, so the two never share a connection pool. That
// transport is otelhttp-instrumented via the global tracer/meter providers.
//
// Every Token() call hits /refresh. The returned token's Expiry comes from the
// JWT exp claim, so callers wanting caching can wrap with
// [oauth2.ReuseTokenSource].
//
// Returns a *[ConfigError] if oauthHost or refreshToken is empty.
func New(oauthHost, refreshToken string, opts ...Option) (*TokenSource, error) {
	if refreshToken == "" {
		return nil, &restkit.ConfigError{
			Name:   clientName,
			Reason: "empty refreshToken",
		}
	}
	var cfg config
	for _, opt := range opts {
		opt(&cfg)
	}
	rk, err := restkit.New(
		oauthHost,
		restkit.WithName(clientName),
		restkit.WithHTTPClient(cfg.httpClient),
	)
	if err != nil {
		return nil, err
	}
	return &TokenSource{
		rkClient:          rk,
		refreshToken:      refreshToken,
		allowedPortfolios: cfg.allowedPortfolios,
	}, nil
}

// Option configures a [TokenSource] at construction time. See
// [WithHTTPClient] and [WithAllowedPortfolios].
type Option func(*config)

// config carries construction-time settings an Option mutates before New builds
// the immutable TokenSource.
type config struct {
	httpClient        *http.Client
	allowedPortfolios []string
}

// WithHTTPClient overrides the default *http.Client used to POST /refresh (for
// a custom Timeout, Transport, proxy, etc.). A nil client is ignored, so [New]
// falls back to restkit's default (30s Timeout, stdlib Transport wrapped with
// otelhttp).
//
// The supplied client MUST NOT be wrapped with an oauth2.Transport backed by
// this same source — that causes infinite recursion on refresh.
func WithHTTPClient(h *http.Client) Option {
	return func(c *config) { c.httpClient = h }
}

// WithAllowedPortfolios scopes the issued JWT to the given portfolio IDs via an
// `allowedPortfolios` array in the /refresh body. Pass nothing, nil, or an empty
// slice to omit the field and let Alor return a token covering all portfolios
// accessible to the refresh token (the default).
//
// The slice is cloned, so a caller writing to ps after WithAllowedPortfolios(ps...)
// cannot mutate the TokenSource — preserving its immutability and
// concurrency-safety.
func WithAllowedPortfolios(portfolios ...string) Option {
	return func(c *config) { c.allowedPortfolios = slices.Clone(portfolios) }
}

// TokenSource is an [oauth2.TokenSource] backed by Alor's non-RFC /refresh
// endpoint. Construct via [New]; immutable after construction and concurrency-
// safe. The refreshToken is kept on the source (never in the restkit base URL)
// so the endpoint can be logged without leaking it.
type TokenSource struct {
	rkClient          *restkit.Client
	refreshToken      string
	allowedPortfolios []string
}

// Token implements [oauth2.TokenSource]. A non-2xx /refresh response is a
// *[ResponseError] (so callers can match 401/403 = bad/expired refresh token);
// every other failure is a *[RequestError] whose Op names the stage — "send" for
// a transport failure (a *url.Error stays in the chain), "unmarshal" for a 2xx
// body that is not a usable token (malformed JSON, empty/non-JWT AccessToken).
// The cause stays reachable via errors.As/errors.Is.
//
// The oauth2.TokenSource interface has no context parameter, so caller
// cancellation cannot reach this call; the request is bounded only by the
// *http.Client's Timeout (restkit's default, or one set via [WithHTTPClient]).
func (s *TokenSource) Token() (*oauth2.Token, error) {
	tb := struct {
		Token             string   `json:"token"`
		AllowedPortfolios []string `json:"allowedPortfolios,omitempty"`
	}{
		Token:             s.refreshToken,
		AllowedPortfolios: s.allowedPortfolios,
	}

	payload, err := restkit.Do[struct {
		AccessToken string `json:"AccessToken"`
	}](context.Background(), s.rkClient, http.MethodPost, refreshPath, tb)
	if err != nil {
		return nil, err
	}
	if payload.AccessToken == "" {
		return nil, badToken(errors.New("empty AccessToken"))
	}

	exp, err := jwtExpiry(payload.AccessToken)
	if err != nil {
		return nil, badToken(err)
	}

	return &oauth2.Token{
		AccessToken: payload.AccessToken,
		TokenType:   "Bearer",
		Expiry:      exp,
	}, nil
}

// badToken tags a 2xx-but-unusable-token failure as an OpUnmarshal
// *[RequestError], matching restkit's decode-stage error so callers see one
// consistent category for "response body was not a usable token".
func badToken(cause error) error {
	return &restkit.RequestError{
		Name: clientName,
		Op:   OpUnmarshal,
		Err:  cause,
	}
}

// jwtExpiry decodes the JWT exp claim so callers wrapping in
// oauth2.ReuseTokenSource get accurate cache invalidation.
func jwtExpiry(token string) (time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("not a JWT: %d segments", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("decode JWT payload: %w", err)
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Time{}, fmt.Errorf("unmarshal JWT claims: %w", err)
	}
	if claims.Exp == 0 {
		return time.Time{}, errors.New("missing JWT exp claim")
	}
	return time.Unix(claims.Exp, 0), nil
}
