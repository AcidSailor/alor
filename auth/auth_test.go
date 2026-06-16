package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func fakeJWT(exp time.Time) string {
	enc := base64.RawURLEncoding.EncodeToString
	header := enc([]byte(`{"alg":"none","typ":"JWT"}`))
	payload := enc(fmt.Appendf(nil, `{"exp":%d}`, exp.Unix()))
	return header + "." + payload + ".sig"
}

func TestNew_InvalidConfig(t *testing.T) {
	cases := map[string]struct {
		host    string
		token   string
		opts    []Option
		wantSub string
	}{
		"empty_host":  {"", "r", nil, "empty endpoint"},
		"empty_token": {"https://x", "", nil, "empty refreshToken"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ts, err := New(tc.host, tc.token, tc.opts...)
			require.Error(t, err)
			assert.Nil(t, ts, "expected nil TokenSource on error")
			var ce *ConfigError
			require.ErrorAs(t, err, &ce)
			assert.Contains(t, err.Error(), tc.wantSub)
		})
	}
}

func TestWithHTTPClient_NilFallsBackToDefault(t *testing.T) {
	ts, err := New("https://x", "r", WithHTTPClient(nil))
	require.NoError(
		t,
		err,
		"a nil *http.Client falls back to restkit's default",
	)
	assert.NotNil(t, ts)
}

type recordingRT struct {
	calls int
	paths []string
	base  http.RoundTripper
}

func (r *recordingRT) RoundTrip(req *http.Request) (*http.Response, error) {
	r.calls++
	r.paths = append(r.paths, req.URL.Path)
	return r.base.RoundTrip(req)
}

func TestWithHTTPClient_RoutesThroughCustomClient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			_, _ = fmt.Fprintf(
				w, `{"AccessToken":%q}`,
				fakeJWT(time.Now().Add(1*time.Hour)),
			)
		},
	))
	defer srv.Close()

	rt := &recordingRT{base: http.DefaultTransport}
	custom := &http.Client{Transport: rt, Timeout: 5 * time.Second}

	ts, err := New(srv.URL, "r", WithHTTPClient(custom))
	require.NoError(t, err)

	_, err = ts.Token()
	require.NoError(t, err)
	assert.Equal(t, 1, rt.calls, "custom RoundTripper called exactly once")
	assert.Equal(t, []string{"/refresh"}, rt.paths)
}

func TestJwtExpiry_Malformed(t *testing.T) {
	enc := base64.RawURLEncoding.EncodeToString
	cases := map[string]struct {
		token   string
		wantSub string
	}{
		"not_a_jwt":          {"oneonly", "not a JWT"},
		"bad_base64_payload": {"aaa.!!!.ccc", "decode JWT payload"},
		"invalid_json_payload": {
			"aaa." + enc([]byte("not json")) + ".ccc",
			"unmarshal JWT claims",
		},
		"missing_exp": {
			"aaa." + enc([]byte(`{"foo":1}`)) + ".ccc",
			"missing JWT exp claim",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := jwtExpiry(tc.token)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantSub)
		})
	}
}

func TestNew_CallsRefreshAndReturnsAccessToken(t *testing.T) {
	exp := time.Now().Add(30 * time.Minute).Truncate(time.Second)
	jwt := fakeJWT(exp)
	var gotBodyToken, gotContentType, gotRawQuery string
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/refresh", r.URL.Path)
			gotRawQuery = r.URL.RawQuery
			gotContentType = r.Header.Get("Content-Type")
			raw, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var payload struct {
				Token string `json:"token"`
			}
			require.NoError(t, json.Unmarshal(raw, &payload))
			gotBodyToken = payload.Token
			_, _ = fmt.Fprintf(w, `{"AccessToken":%q}`, jwt)
		},
	))
	defer srv.Close()

	ts, err := New(srv.URL, "refresh-xyz")
	require.NoError(t, err)
	tok, err := ts.Token()
	require.NoError(t, err)
	assert.Equal(t, "refresh-xyz", gotBodyToken)
	assert.Equal(t, "application/json", gotContentType)
	assert.Empty(t, gotRawQuery, "token must not be sent in URL query")
	assert.Equal(t, jwt, tok.AccessToken)
	assert.Equal(t, "Bearer", tok.TokenType)
	assert.True(
		t,
		tok.Expiry.Equal(exp),
		"expiry: got %s, want %s",
		tok.Expiry,
		exp,
	)
}

func TestWithAllowedPortfolios(t *testing.T) {
	type body struct {
		Token             string   `json:"token"`
		AllowedPortfolios []string `json:"allowedPortfolios,omitempty"`
	}
	cases := map[string]struct {
		opts []Option
		want []string
	}{
		"unset_omits_field": {
			opts: nil,
			want: nil,
		},
		"empty_slice_omits_field": {
			opts: []Option{WithAllowedPortfolios()},
			want: nil,
		},
		"explicit_ids_sent": {
			opts: []Option{WithAllowedPortfolios("P1", "P2", "P3")},
			want: []string{"P1", "P2", "P3"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var got body
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					raw, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					require.NoError(t, json.Unmarshal(raw, &got))
					_, _ = fmt.Fprintf(
						w, `{"AccessToken":%q}`,
						fakeJWT(time.Now().Add(time.Hour)),
					)
				},
			))
			defer srv.Close()

			ts, err := New(srv.URL, "r", tc.opts...)
			require.NoError(t, err)
			_, err = ts.Token()
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.AllowedPortfolios)
		})
	}
}

func TestNew_NonJWTAccessTokenErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"AccessToken":"not-a-jwt"}`))
		},
	))
	defer srv.Close()

	ts, err := New(srv.URL, "r")
	require.NoError(t, err)
	_, err = ts.Token()
	require.Error(t, err)
	var re *RequestError
	require.ErrorAs(t, err, &re)
	assert.Equal(t, "unmarshal", re.Op)
	assert.Contains(t, err.Error(), "JWT")
}

func TestNew_PropagatesHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("bad refresh"))
		},
	))
	defer srv.Close()

	ts, err := New(srv.URL, "expired")
	require.NoError(t, err)
	_, err = ts.Token()
	require.Error(t, err)
	var re *ResponseError
	require.ErrorAs(t, err, &re)
	assert.Equal(t, http.StatusForbidden, re.StatusCode)
	assert.Contains(t, re.Body, "bad refresh")
}

func TestTokenSource_MalformedResponse(t *testing.T) {
	cases := map[string]struct {
		body    string
		wantSub string
	}{
		"invalid_json":      {`{not json`, "invalid character"},
		"empty_accesstoken": {`{"AccessToken":""}`, "empty"},
		"html_body": {
			`<html><body>oops</body></html>`,
			"invalid character",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte(tc.body))
				},
			))
			defer srv.Close()

			ts, err := New(srv.URL, "r")
			require.NoError(t, err)
			_, err = ts.Token()
			require.Error(t, err)
			var re *RequestError
			require.ErrorAs(t, err, &re)
			assert.Equal(t, "unmarshal", re.Op)
			assert.Contains(t, err.Error(), tc.wantSub)
		})
	}
}

func TestTokenSource_ReuseTokenSourceCaches(t *testing.T) {
	t.Run("fresh_exp_one_refresh", func(t *testing.T) {
		exp := time.Now().Add(1 * time.Hour).Truncate(time.Second)
		var hits int
		srv := httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				hits++
				_, _ = fmt.Fprintf(w, `{"AccessToken":%q}`, fakeJWT(exp))
			},
		))
		defer srv.Close()

		inner, err := New(srv.URL, "r")
		require.NoError(t, err)
		ts := oauth2.ReuseTokenSource(nil, inner)

		for range 3 {
			_, err := ts.Token()
			require.NoError(t, err)
		}
		assert.Equal(t, 1, hits)
	})

	t.Run("expired_exp_refetches", func(t *testing.T) {
		exp := time.Now().Add(-1 * time.Hour).Truncate(time.Second)
		var hits int
		srv := httptest.NewServer(http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				hits++
				_, _ = fmt.Fprintf(w, `{"AccessToken":%q}`, fakeJWT(exp))
			},
		))
		defer srv.Close()

		inner, err := New(srv.URL, "r")
		require.NoError(t, err)
		ts := oauth2.ReuseTokenSource(nil, inner)

		for range 3 {
			_, err := ts.Token()
			require.NoError(t, err)
		}
		assert.Equal(t, 3, hits)
	})
}
