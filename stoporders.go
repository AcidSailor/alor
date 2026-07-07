package alor

import (
	"context"
	"net/http"
	"strconv"

	"github.com/acidsailor/restkit"
)

// stopOrdersService groups the StopOrders API operations. Obtain it via Client.StopOrders.
type stopOrdersService struct{ c *Client }

// StopOrdersListRequest selects the exchange/portfolio whose stop orders to
// return.
type StopOrdersListRequest struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
}

// List returns all stop orders for the given exchange/portfolio.
func (s *stopOrdersService) List(
	ctx context.Context,
	params StopOrdersListRequest,
) (*ResponseStopOrdersWarpHeavy, error) {
	return do[*ResponseStopOrdersWarpHeavy](ctx, s.c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/stoporders"),
		heavyValues(), nil)
}

// StopOrdersGetRequest selects the single stop order to return.
type StopOrdersGetRequest struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
	OrderID   int64  `json:"orderId"`
}

// Get returns a single stop order by id for the given
// exchange/portfolio.
func (s *stopOrdersService) Get(
	ctx context.Context,
	params StopOrdersGetRequest,
) (*ResponseStopOrderWarpHeavy, error) {
	path := clientPath(params.Exchange, params.Portfolio,
		"/stoporders/"+strconv.FormatInt(params.OrderID, 10))
	return do[*ResponseStopOrderWarpHeavy](ctx, s.c, http.MethodGet,
		path, heavyValues(), nil)
}

// StopOrdersPlaceRequest carries the idempotency key and the stop-order body.
type StopOrdersPlaceRequest struct {
	ReqID string                        `json:"-"`
	Order OrdersActionsStopMarketTVWarp `json:"order"`
}

// Place submits a new stop (stop-market) order.
func (s *stopOrdersService) Place(
	ctx context.Context,
	params StopOrdersPlaceRequest,
) (*ResponseOrderActionLimitMarketCommandAPI, error) {
	return do[*ResponseOrderActionLimitMarketCommandAPI](
		ctx,
		s.c,
		http.MethodPost,
		ordersBase+"/actions/stop",
		restkit.NewValues(),
		params.Order,
		withReqID(params.ReqID),
	)
}

// StopOrdersPlaceLimitRequest carries the idempotency key and the stop-limit-order
// body.
type StopOrdersPlaceLimitRequest struct {
	ReqID string                           `json:"-"`
	Order OrdersActionsStopLimitTVWarpPost `json:"order"`
}

// PlaceLimit submits a new stop-limit order.
func (s *stopOrdersService) PlaceLimit(
	ctx context.Context,
	params StopOrdersPlaceLimitRequest,
) (*ResponseOrderActionLimitMarketCommandAPI, error) {
	return do[*ResponseOrderActionLimitMarketCommandAPI](
		ctx,
		s.c,
		http.MethodPost,
		ordersBase+"/actions/stopLimit",
		restkit.NewValues(),
		params.Order,
		withReqID(params.ReqID),
	)
}

// StopOrdersReplaceRequest carries the idempotency key, the target stop-order id,
// and the replacement stop-order body.
type StopOrdersReplaceRequest struct {
	ReqID       string                        `json:"-"`
	StopOrderID int64                         `json:"stopOrderId"` // path
	Order       OrdersActionsStopMarketTVWarp `json:"order"`
}

// Replace replaces the stop order with the given id.
func (s *stopOrdersService) Replace(
	ctx context.Context,
	params StopOrdersReplaceRequest,
) (*ResponseOrderActionLimitMarketCommandAPI, error) {
	path := ordersBase + "/actions/stop/" + itoa64(params.StopOrderID)
	return do[*ResponseOrderActionLimitMarketCommandAPI](
		ctx,
		s.c,
		http.MethodPut,
		path,
		restkit.NewValues(),
		params.Order,
		withReqID(params.ReqID),
	)
}

// StopOrdersReplaceLimitRequest carries the idempotency key, the target
// stop-order id, and the replacement stop-limit-order body.
type StopOrdersReplaceLimitRequest struct {
	ReqID       string                          `json:"-"`
	StopOrderID int64                           `json:"stopOrderId"` // path
	Order       OrdersActionsStopLimitTVWarpPut `json:"order"`
}

// ReplaceLimit replaces the stop-limit order with the given id.
func (s *stopOrdersService) ReplaceLimit(
	ctx context.Context,
	params StopOrdersReplaceLimitRequest,
) (*ResponseOrderActionLimitMarketCommandAPI, error) {
	path := ordersBase + "/actions/stopLimit/" + itoa64(params.StopOrderID)
	return do[*ResponseOrderActionLimitMarketCommandAPI](
		ctx,
		s.c,
		http.MethodPut,
		path,
		restkit.NewValues(),
		params.Order,
		withReqID(params.ReqID),
	)
}
