package alor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetOrderBookForwardsOptions verifies path assembly, the pinned Heavy
// format, and that optional fields (Depth, InstrumentGroup) reach the query.
func TestGetOrderBookForwardsOptions(t *testing.T) {
	var gotPath string
	var gotQuery url.Values
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotQuery = r.URL.Query()
			_, _ = w.Write([]byte("{}"))
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, staticTS())
	require.NoError(t, err)

	_, err = c.MarketData.OrderBook(
		context.Background(),
		MarketDataOrderBookRequest{
			Exchange:        "MOEX",
			Symbol:          "SBER",
			Depth:           new(20),
			InstrumentGroup: new("TQBR"),
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "/md/v2/orderbooks/MOEX/SBER", gotPath)
	assert.Equal(t, "Heavy", gotQuery.Get("format"))
	assert.Equal(t, "20", gotQuery.Get("depth"))
	assert.Equal(t, "TQBR", gotQuery.Get("instrumentGroup"))
}

// TestGetHistorySetsRequiredParams verifies the required candle params and an
// optional field are forwarded alongside the pinned Heavy format.
func TestGetHistorySetsRequiredParams(t *testing.T) {
	var q url.Values
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q = r.URL.Query()
			_, _ = w.Write([]byte("{}"))
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, staticTS())
	require.NoError(t, err)

	_, err = c.MarketData.History(
		context.Background(),
		MarketDataHistoryRequest{
			Exchange:  "MOEX",
			Symbol:    "SBER",
			Tf:        "60",
			From:      1700000000,
			To:        1700003600,
			CountBack: new(10),
		},
	)
	require.NoError(t, err)
	assert.Equal(t, "SBER", q.Get("symbol"))
	assert.Equal(t, "MOEX", q.Get("exchange"))
	assert.Equal(t, "60", q.Get("tf"))
	assert.Equal(t, "1700000000", q.Get("from"))
	assert.Equal(t, "1700003600", q.Get("to"))
	assert.Equal(t, "10", q.Get("countBack"))
	assert.Equal(t, "Heavy", q.Get("format"))
}

// TestPlaceLimitOrderEmptyReqIDOmitsHeader verifies the withReqID no-op: an
// empty reqID leaves X-REQID off the wire rather than sending a blank one.
func TestPlaceLimitOrderEmptyReqIDOmitsHeader(t *testing.T) {
	var reqIDValues []string
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqIDValues = r.Header.Values("X-REQID")
			_, _ = w.Write([]byte(`{"orderNumber":"1"}`))
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, staticTS())
	require.NoError(t, err)

	_, err = c.Orders.PlaceLimit(
		context.Background(),
		OrdersPlaceLimitRequest{Order: OrdersActionsLimitTVPost{}},
	)
	require.NoError(t, err)
	assert.Empty(t, reqIDValues, "empty reqID omits the X-REQID header")
}

// TestGetAvailableBoardsOmitsFormat: a non-_Heavy endpoint does not pin format.
func TestGetAvailableBoardsOmitsFormat(t *testing.T) {
	var hasFormat bool
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, hasFormat = r.URL.Query()["format"]
			_, _ = w.Write([]byte("[]"))
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, staticTS())
	require.NoError(t, err)

	_, err = c.MarketData.Boards(context.Background(),
		MarketDataBoardsRequest{Exchange: "MOEX", Symbol: "SBER"})
	require.NoError(t, err)
	assert.False(t, hasFormat, "non-Heavy endpoint should not send format")
}

// TestPlaceLimitOrderSendsReqIDAndBody verifies the order-command path, the
// X-REQID header, the JSON content type, and that the response decodes into the
// returned struct.
func TestPlaceLimitOrderSendsReqIDAndBody(t *testing.T) {
	var gotMethod, gotPath, gotReqID, gotContentType, gotBody string
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			gotReqID = r.Header.Get("X-REQID")
			gotContentType = r.Header.Get("Content-Type")
			buf := make([]byte, r.ContentLength)
			_, _ = r.Body.Read(buf)
			gotBody = string(buf)
			_, _ = w.Write([]byte(`{"orderNumber":"18995978560"}`))
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, staticTS())
	require.NoError(t, err)

	resp, err := c.Orders.PlaceLimit(
		context.Background(),
		OrdersPlaceLimitRequest{
			ReqID: "req-123",
			Order: OrdersActionsLimitTVPost{},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t,
		"/commandapi/warptrans/TRADE/v2/client/orders/actions/limit", gotPath)
	assert.Equal(t, "req-123", gotReqID)
	assert.Equal(t, "application/json", gotContentType)
	assert.Contains(t, gotBody, "{")
	require.NotNil(t, resp.OrderNumber)
	assert.Equal(t, "18995978560", *resp.OrderNumber)
}

// TestCancelOrderNoContentPath verifies the DELETE path and required query
// params, and that a text/plain "success" body does not surface as a decode
// error on the no-content (exec) path.
func TestCancelOrderNoContentPath(t *testing.T) {
	var gotMethod, gotPath string
	var q url.Values
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			q = r.URL.Query()
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("success"))
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, staticTS())
	require.NoError(t, err)

	err = c.Orders.Cancel(context.Background(), OrdersCancelRequest{
		OrderID:   18995978560,
		Exchange:  "MOEX",
		Portfolio: "D12345",
		Stop:      false,
	})
	require.NoError(t, err)
	assert.Equal(t, http.MethodDelete, gotMethod)
	assert.Equal(t,
		"/commandapi/warptrans/TRADE/v2/client/orders/18995978560", gotPath)
	assert.Equal(t, "MOEX", q.Get("exchange"))
	assert.Equal(t, "D12345", q.Get("portfolio"))
	assert.Equal(t, "false", q.Get("stop"))
}

// TestMutationErrorWrapping verifies the mutating path honors the same error
// contract as reads: a non-2xx surfaces as a *ResponseError. Also guards the
// exec discard path — a non-2xx must not be swallowed as an expected unmarshal
// failure.
func TestMutationErrorWrapping(t *testing.T) {
	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "bad order", http.StatusBadRequest)
		}),
	)
	defer srv.Close()

	c, err := NewClient(srv.URL, staticTS())
	require.NoError(t, err)

	err = c.OrderGroups.Delete(context.Background(),
		OrderGroupsDeleteRequest{OrderGroupID: "abc"})
	var respErr *ResponseError
	require.ErrorAs(t, err, &respErr)
	assert.Equal(t, http.StatusBadRequest, respErr.StatusCode)
}
