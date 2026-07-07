// Package alor is a Go client for the Alor broker API. Response models are
// generated from the OpenAPI spec by oapi-codegen (models.gen.go) as plain
// structs; the HTTP transport is the shared restkit core (auth, query, and the
// X-REQID idempotency header ride on restkit request hooks), and the curated
// facade methods live in wrappers.go.
package alor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/acidsailor/restkit"
	"golang.org/x/oauth2"
)

const (
	EndpointProduction = "https://api.alor.ru"
	EndpointTest       = "https://apidev.alor.ru"

	// clientName is carried into every restkit error's Name field so a program
	// using several restkit-based clients can attribute a failure to alor.
	clientName = "alor"

	// formatHeavy is the ?format= value the facade pins on every endpoint whose
	// response oneOf the overlay collapsed to its _Heavy variant, so the wire
	// payload matches the generated Response_*_Heavy structs.
	formatHeavy = "Heavy"
)

// Client is the Alor API client. Operations are methods on the per-service facades
// (client.Orders, client.MarketData, ...); each authenticates with a bearer token and returns a non-2xx
// as a *[ResponseError] or any other per-call failure as a *[RequestError]
// (cause preserved).
//
// A Client is immutable after construction and safe for concurrent use.
//
// OTel instrumentation flows through the global providers
// (otel.GetTracerProvider, otel.GetMeterProvider): the restkit transport is
// wrapped with otelhttp, emitting client spans and the standard HTTP client
// metrics. Callers wire it by setting the globals.
type Client struct {
	rkClient *restkit.Client

	Orders      *ordersService
	StopOrders  *stopOrdersService
	OrderGroups *orderGroupsService
	Portfolio   *portfolioService
	Trades      *tradesService
	MarketData  *marketDataService
}

// Option configures a Client at construction time. See [WithHTTPClient] and
// [WithInitialToken].
type Option func(*config)

// config carries construction-time settings an Option mutates before NewClient
// builds the immutable Client.
type config struct {
	httpClient   *http.Client
	initialToken *oauth2.Token
}

// WithHTTPClient overrides the default *http.Client for outgoing API calls —
// for a custom Timeout, Transport, proxy, etc. A nil client is ignored, so
// [NewClient] falls back to restkit's default (30s Timeout, stdlib Transport).
//
// NewClient wraps the client's Transport with otelhttp; the caller's client is
// copied, not mutated.
func WithHTTPClient(h *http.Client) Option {
	return func(c *config) { c.httpClient = h }
}

// WithInitialToken seeds the internal [oauth2.ReuseTokenSource] cache with t, so
// the first request reuses t instead of refreshing — useful for resuming from a
// persisted token across restarts. t is reused until its Expiry; once stale, the
// underlying TokenSource is called as normal.
//
// A nil t is a no-op (cache starts empty, first request refreshes).
func WithInitialToken(t *oauth2.Token) Option {
	return func(c *config) { c.initialToken = t }
}

// NewClient builds a Client that authenticates every request with a bearer
// token from ts, applying any options. Use [EndpointProduction] or
// [EndpointTest] for endpoint. For ts, pass auth.New for Alor's auto-refresh
// flow, or any other oauth2.TokenSource.
//
// ts is wrapped with [oauth2.ReuseTokenSource], so tokens are cached until their
// Expiry and concurrent refreshes are serialized. No network calls happen during
// construction; the first refresh happens on the first request — to skip it
// (e.g. resuming from a persisted token), seed the cache via [WithInitialToken].
//
// Returns a *[ConfigError] if ts is nil or endpoint is empty.
func NewClient(
	endpoint string,
	ts oauth2.TokenSource,
	opts ...Option,
) (*Client, error) {
	if ts == nil {
		return nil, &restkit.ConfigError{
			Name:   clientName,
			Reason: "nil TokenSource",
		}
	}
	var cfg config
	for _, opt := range opts {
		opt(&cfg)
	}

	rkClient, err := restkit.New(
		endpoint,
		restkit.WithName(clientName),
		restkit.WithHTTPClient(cfg.httpClient),
		restkit.WithHook(
			bearerAuth(oauth2.ReuseTokenSource(cfg.initialToken, ts)),
		),
	)
	if err != nil {
		return nil, err
	}
	c := &Client{rkClient: rkClient}
	c.Orders = &ordersService{c}
	c.StopOrders = &stopOrdersService{c}
	c.OrderGroups = &orderGroupsService{c}
	c.Portfolio = &portfolioService{c}
	c.Trades = &tradesService{c}
	c.MarketData = &marketDataService{c}
	return c, nil
}

// bearerAuth returns the client-wide hook that stamps each request with a fresh
// bearer token from ts. A token-source failure aborts the call as a
// *[RequestError] with Op == [restkit.OpHook], cause preserved and reachable via
// errors.As/errors.Is — for an auth.TokenSource, that cause is itself a restkit
// *ResponseError (a non-2xx /refresh) or *RequestError.
func bearerAuth(ts oauth2.TokenSource) restkit.RequestHook {
	return func(r *http.Request) error {
		tok, err := ts.Token()
		if err != nil {
			return err
		}
		r.Header.Set("Authorization", "Bearer "+tok.AccessToken)
		return nil
	}
}

// withReqID returns a per-call hook that sets the X-REQID idempotency header the
// order-command endpoints require. An empty id is a no-op (header omitted).
func withReqID(id string) restkit.RequestHook {
	return func(r *http.Request) error {
		if id != "" {
			r.Header.Set("X-REQID", id)
		}
		return nil
	}
}

// do issues a request through restkit, merging q into the query, and decodes the
// 2xx JSON body into T. A non-2xx surfaces as a *[ResponseError]; any other
// per-call failure (auth, encode, build, transport, decode) as a *[RequestError]
// (cause preserved).
func do[T any](
	ctx context.Context,
	c *Client,
	method, path string,
	q restkit.Values,
	body any,
	hooks ...restkit.RequestHook,
) (T, error) {
	h := append([]restkit.RequestHook{restkit.WithQuery(q.Values)}, hooks...)
	return restkit.Do[T](ctx, c.rkClient, method, path, body, h...)
}

// exec issues a request through restkit and discards the 2xx body. It backs the
// endpoints that reply with a bare text/plain "success" and no actionable
// payload (order cancellation, order-group modify/delete).
//
// restkit always JSON-decodes a 2xx body, so the non-JSON body fails to
// unmarshal; that expected *[RequestError] (Op == [restkit.OpUnmarshal]) is
// treated as success. A non-2xx is still reported as a *[ResponseError] (restkit
// checks status before unmarshalling), and any other failure keeps its Op.
//
// Failure is assumed to come only via a non-2xx status: any 2xx is success
// regardless of body, so exec does not guard against a hypothetical
// 2xx-with-failure-envelope contract change.
//
// The check is a concrete top-level type assertion, not errors.As: restkit
// returns the OpUnmarshal failure as the outermost error, whereas an auth
// failure via the bearer hook is an OpHook *RequestError merely *wrapping* an
// OpUnmarshal cause (auth.badToken). errors.As would walk into that cause and
// wrongly swallow a failed order command as success; the assertion matches only
// the genuine non-JSON-body case.
func exec(
	ctx context.Context,
	c *Client,
	method, path string,
	q restkit.Values,
	body any,
	hooks ...restkit.RequestHook,
) error {
	_, err := do[json.RawMessage](ctx, c, method, path, q, body, hooks...)
	if re, ok := err.(*restkit.RequestError); ok &&
		re.Op == OpUnmarshal {
		return nil
	}
	return err
}

// clientPath builds the /md/v2/Clients/{exchange}/{portfolio}/<suffix> path,
// escaping the path parameters.
func clientPath(exchange, portfolio, suffix string) string {
	return "/md/v2/Clients/" + url.PathEscape(exchange) +
		"/" + url.PathEscape(portfolio) + suffix
}

// ServerTime returns Alor's current server time as a Unix timestamp (seconds).
func (c *Client) ServerTime(ctx context.Context) (int64, error) {
	return do[int64](ctx, c, http.MethodGet, "/md/v2/time",
		restkit.NewValues(), nil)
}
