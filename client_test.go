package alor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/acidsailor/restkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// staticTS is a trivial oauth2.TokenSource for tests: never refreshes, never
// errors, so it exercises the facade without auth network traffic.
func staticTS() oauth2.TokenSource {
	return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"})
}

// errTokenSource is an oauth2.TokenSource whose Token() always fails with a
// fixed error, to exercise the bearer-auth hook failure path and the resulting
// cross-package error composition.
type errTokenSource struct{ err error }

func (e errTokenSource) Token() (*oauth2.Token, error) { return nil, e.err }

func TestNewClientValidation(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		ts       oauth2.TokenSource
		wantErr  bool
	}{
		{"empty endpoint", "", staticTS(), true},
		{"nil token source", EndpointTest, nil, true},
		{"valid", EndpointTest, staticTS(), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewClient(tc.endpoint, tc.ts)
			if tc.wantErr {
				var ce *ConfigError
				require.ErrorAs(t, err, &ce)
				assert.Nil(t, c, "expected nil client on error")
				return
			}
			require.NoError(t, err)
			require.NotNil(t, c)
		})
	}
}

// TestNewClientTrimsTrailingSlash: the endpoint is normalized so a trailing
// slash does not produce a doubled slash in the request path.
func TestNewClientTrimsTrailingSlash(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			_, _ = w.Write([]byte("1700000000"))
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL+"/", staticTS())
	require.NoError(t, err)

	_, err = c.ServerTime(context.Background())
	require.NoError(t, err)
	assert.Equal(
		t,
		"/md/v2/time",
		gotPath,
		"no doubled slash from trailing-slash endpoint",
	)
}

// TestNewClientNilHTTPClientOption: WithHTTPClient(nil) is a no-op; restkit
// falls back to its default *http.Client rather than erroring.
func TestNewClientNilHTTPClientOption(t *testing.T) {
	c, err := NewClient(EndpointTest, staticTS(), WithHTTPClient(nil))
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestNewClientOptionsApplied(t *testing.T) {
	custom := &http.Client{Timeout: time.Second}
	c, err := NewClient(
		EndpointTest, staticTS(),
		WithHTTPClient(custom),
		WithInitialToken(&oauth2.Token{AccessToken: "seed"}),
	)
	require.NoError(t, err)
	require.NotNil(t, c)
}

// TestNewClientNilInitialToken: WithInitialToken(nil) is a no-op, equivalent to
// omitting the option, rather than an error.
func TestNewClientNilInitialToken(t *testing.T) {
	c, err := NewClient(EndpointTest, staticTS(), WithInitialToken(nil))
	require.NoError(t, err)
	require.NotNil(t, c)
}

// TestServerTimeSuccess exercises the happy path end-to-end: the request hits
// the expected path with the bearer token attached, and the body decodes into
// the returned value.
func TestServerTimeSuccess(t *testing.T) {
	var gotPath, gotAuth string
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotAuth = r.Header.Get("Authorization")
			_, _ = w.Write([]byte("1700000000"))
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, staticTS())
	require.NoError(t, err)

	ts, err := c.ServerTime(context.Background())
	require.NoError(t, err)
	assert.Equal(t, int64(1700000000), ts)
	assert.Equal(t, "/md/v2/time", gotPath)
	assert.Equal(t, "Bearer test-token", gotAuth)
}

// TestGetOrdersPinsHeavyFormat verifies the Clients path assembly and that the
// facade pins ?format=Heavy on a _Heavy-collapsed endpoint.
func TestGetOrdersPinsHeavyFormat(t *testing.T) {
	var gotPath, gotFormat string
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotFormat = r.URL.Query().Get("format")
			_, _ = w.Write([]byte("[]"))
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, staticTS())
	require.NoError(t, err)

	orders, err := c.Orders.List(context.Background(),
		OrdersListRequest{Exchange: "MOEX", Portfolio: "D12345"})
	require.NoError(t, err)
	assert.Empty(t, orders)
	assert.Equal(t, "/md/v2/Clients/MOEX/D12345/orders", gotPath)
	assert.Equal(t, "Heavy", gotFormat)
}

// TestSearchSecuritiesQuery verifies the optional free-text query is forwarded
// alongside the pinned format.
func TestSearchSecuritiesQuery(t *testing.T) {
	var gotQuery, gotFormat string
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotQuery = r.URL.Query().Get("query")
			gotFormat = r.URL.Query().Get("format")
			_, _ = w.Write([]byte("[]"))
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, staticTS())
	require.NoError(t, err)

	_, err = c.MarketData.Search(context.Background(),
		MarketDataSearchRequest{Query: new("SBER")})
	require.NoError(t, err)
	assert.Equal(t, "SBER", gotQuery)
	assert.Equal(t, "Heavy", gotFormat)
}

// TestErrorWrapping verifies the facade's error contract: a non-2xx surfaces as
// a *ResponseError carrying the status code and raw body, and — being the typed
// result of the call — is not a *RequestError.
func TestErrorWrapping(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "boom", http.StatusInternalServerError)
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, staticTS())
	require.NoError(t, err)

	_, err = c.ServerTime(context.Background())

	var respErr *ResponseError
	require.ErrorAs(t, err, &respErr)
	assert.Equal(t, http.StatusInternalServerError, respErr.StatusCode)
	assert.Contains(t, respErr.Body, "boom")
	var reqErr *RequestError
	assert.NotErrorAs(t, err, &reqErr, "a non-2xx is not a *RequestError")
}

// TestRequestErrorWrapping verifies a non-ResponseError failure (here a decode
// error on a malformed 2xx body) surfaces as a *RequestError while the cause
// stays reachable.
func TestRequestErrorWrapping(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`"not-an-int"`))
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, staticTS())
	require.NoError(t, err)

	_, err = c.ServerTime(context.Background())
	var reqErr *RequestError
	require.ErrorAs(t, err, &reqErr)

	var respErr *ResponseError
	assert.NotErrorAs(
		t,
		err,
		&respErr,
		"a decode failure is not a *ResponseError",
	)
}

// TestBearerAuthHookFailurePropagatesCause verifies the cross-package error
// contract the design rests on: a token-source failure aborts the call as a
// *RequestError tagged OpHook, the request is never sent, and the auth package's
// typed cause (here a *ResponseError carrying a 403 — expired refresh token)
// stays reachable via errors.As so callers can re-auth on 401/403.
func TestBearerAuthHookFailurePropagatesCause(t *testing.T) {
	var hit bool
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			hit = true
			_, _ = w.Write([]byte("1700000000"))
		}),
	)
	defer srv.Close()

	authErr := &ResponseError{
		Name:       "alor-auth",
		StatusCode: http.StatusForbidden,
		Body:       "expired refresh token",
	}
	c, err := NewClient(srv.URL, errTokenSource{err: authErr})
	require.NoError(t, err)

	_, err = c.ServerTime(context.Background())

	var reqErr *RequestError
	require.ErrorAs(t, err, &reqErr)
	assert.Equal(t, restkit.OpHook, reqErr.Op)

	var respErr *ResponseError
	require.ErrorAs(t, err, &respErr, "auth cause stays reachable")
	assert.Equal(t, http.StatusForbidden, respErr.StatusCode)
	assert.False(t, hit, "no request is sent when the auth hook fails")
}

// TestExecDoesNotSwallowAuthFailure guards the hardened exec swallow: an auth
// failure on a bare-"success" endpoint surfaces as an OpHook *RequestError that
// merely wraps an OpUnmarshal-tagged cause (mirroring auth.badToken). exec must
// not mistake that wrapped OpUnmarshal for the expected non-JSON-body case and
// report a failed cancellation as success.
func TestExecDoesNotSwallowAuthFailure(t *testing.T) {
	var hit bool
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			hit = true
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("success"))
		}),
	)
	defer srv.Close()

	// An OpUnmarshal *RequestError is what exec swallows as the top-level error;
	// here it is the auth cause wrapped under OpHook.
	authErr := &RequestError{
		Name: "alor-auth",
		Op:   OpUnmarshal,
		Err:  errors.New("empty AccessToken"),
	}
	c, err := NewClient(srv.URL, errTokenSource{err: authErr})
	require.NoError(t, err)

	err = c.Orders.Cancel(context.Background(), OrdersCancelRequest{
		OrderID:   1,
		Exchange:  "MOEX",
		Portfolio: "D12345",
		Stop:      false,
	})
	require.Error(t, err, "auth failure must not be swallowed as success")

	var reqErr *RequestError
	require.ErrorAs(t, err, &reqErr)
	assert.Equal(t, OpHook, reqErr.Op)
	assert.False(t, hit, "no request is sent when the auth hook fails")
}
