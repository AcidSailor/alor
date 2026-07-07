package alor

import (
	"context"
	"net/http"
	"strconv"

	"github.com/acidsailor/restkit"
)

// ordersService groups the Orders API operations. Obtain it via Client.Orders.
type ordersService struct{ c *Client }

// ordersBase is the common prefix for the CommandAPI order-action endpoints.
const ordersBase = "/commandapi/warptrans/TRADE/v2/client/orders"

// OrdersListRequest selects the exchange/portfolio whose orders to return.
type OrdersListRequest struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
}

// List returns all orders for the given exchange/portfolio.
func (s *ordersService) List(
	ctx context.Context,
	params OrdersListRequest,
) (*ResponseOrdersHeavy, error) {
	return do[*ResponseOrdersHeavy](ctx, s.c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/orders"),
		heavyValues(), nil)
}

// OrdersGetRequest selects the single order to return.
type OrdersGetRequest struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
	OrderID   int64  `json:"orderId"`
}

// Get returns a single order by id for the given exchange/portfolio.
func (s *ordersService) Get(
	ctx context.Context,
	params OrdersGetRequest,
) (*ResponseOrderHeavy, error) {
	path := clientPath(params.Exchange, params.Portfolio,
		"/orders/"+strconv.FormatInt(params.OrderID, 10))
	return do[*ResponseOrderHeavy](ctx, s.c, http.MethodGet,
		path, heavyValues(), nil)
}

// OrdersPlaceMarketRequest carries the idempotency key and the market-order body.
type OrdersPlaceMarketRequest struct {
	ReqID string                `json:"-"`     // X-REQID idempotency key, caller-minted
	Order OrdersActionsMarketTV `json:"order"` // request body
}

// PlaceMarket submits a new market order.
func (s *ordersService) PlaceMarket(
	ctx context.Context,
	params OrdersPlaceMarketRequest,
) (*ResponseOrderActionLimitMarketCommandAPI, error) {
	return do[*ResponseOrderActionLimitMarketCommandAPI](
		ctx,
		s.c,
		http.MethodPost,
		ordersBase+"/actions/market",
		restkit.NewValues(),
		params.Order,
		withReqID(params.ReqID),
	)
}

// OrdersPlaceLimitRequest carries the idempotency key and the limit-order body.
type OrdersPlaceLimitRequest struct {
	ReqID string                   `json:"-"`
	Order OrdersActionsLimitTVPost `json:"order"`
}

// PlaceLimit submits a new limit order.
func (s *ordersService) PlaceLimit(
	ctx context.Context,
	params OrdersPlaceLimitRequest,
) (*ResponseOrderActionLimitMarketCommandAPI, error) {
	return do[*ResponseOrderActionLimitMarketCommandAPI](
		ctx,
		s.c,
		http.MethodPost,
		ordersBase+"/actions/limit",
		restkit.NewValues(),
		params.Order,
		withReqID(params.ReqID),
	)
}

// OrdersReplaceMarketRequest carries the idempotency key, the target order id, and
// the replacement market-order body.
type OrdersReplaceMarketRequest struct {
	ReqID   string                `json:"-"`
	OrderID int64                 `json:"orderId"` // path
	Order   OrdersActionsMarketTV `json:"order"`
}

// ReplaceMarket replaces the market order with the given id.
func (s *ordersService) ReplaceMarket(
	ctx context.Context,
	params OrdersReplaceMarketRequest,
) (*ResponseOrderActionLimitMarket, error) {
	path := ordersBase + "/actions/market/" + itoa64(params.OrderID)
	return do[*ResponseOrderActionLimitMarket](ctx, s.c, http.MethodPut,
		path, restkit.NewValues(), params.Order, withReqID(params.ReqID))
}

// OrdersReplaceLimitRequest carries the idempotency key, the target order id, and
// the replacement limit-order body.
type OrdersReplaceLimitRequest struct {
	ReqID   string                  `json:"-"`
	OrderID int64                   `json:"orderId"` // path
	Order   OrdersActionsLimitTVPut `json:"order"`
}

// ReplaceLimit replaces the limit order with the given id.
func (s *ordersService) ReplaceLimit(
	ctx context.Context,
	params OrdersReplaceLimitRequest,
) (*ResponseOrderActionLimitMarket, error) {
	path := ordersBase + "/actions/limit/" + itoa64(params.OrderID)
	return do[*ResponseOrderActionLimitMarket](ctx, s.c, http.MethodPut,
		path, restkit.NewValues(), params.Order, withReqID(params.ReqID))
}

// OrdersEstimateRequest carries the order to estimate.
type OrdersEstimateRequest struct {
	Order EstimateOrder `json:"order"`
}

// Estimate estimates the cost and lot count of a prospective order without
// submitting it.
func (s *ordersService) Estimate(
	ctx context.Context,
	params OrdersEstimateRequest,
) (*ResponseEstimateOrder, error) {
	return do[*ResponseEstimateOrder](ctx, s.c, http.MethodPost,
		ordersBase+"/estimate", restkit.NewValues(), params.Order)
}

// OrdersEstimateBatchRequest carries the orders to estimate.
type OrdersEstimateBatchRequest struct {
	Orders EstimateOrders `json:"orders"`
}

// EstimateBatch estimates the cost and lot count of several prospective orders
// in one call.
func (s *ordersService) EstimateBatch(
	ctx context.Context,
	params OrdersEstimateBatchRequest,
) (*ResponseEstimateOrders, error) {
	return do[*ResponseEstimateOrders](ctx, s.c, http.MethodPost,
		ordersBase+"/estimate/all", restkit.NewValues(), params.Orders)
}

// OrdersCancelRequest selects the order to cancel and its scope.
type OrdersCancelRequest struct {
	OrderID   int64  `json:"orderId"`   // path
	Exchange  string `json:"exchange"`  // query
	Portfolio string `json:"portfolio"` // query
	Stop      bool   `json:"stop"`      // query; true cancels a stop order
}

// Cancel cancels the order with the given id. Exchange and Portfolio scope
// it; set Stop to true for a stop order. Returns only an error (bare success
// reply).
func (s *ordersService) Cancel(
	ctx context.Context,
	params OrdersCancelRequest,
) error {
	q := restkit.NewValues().
		Str(keyExchange, &params.Exchange).
		Str(keyPortfolio, &params.Portfolio).
		Bool(keyStop, &params.Stop)
	return exec(ctx, s.c, http.MethodDelete,
		ordersBase+"/"+itoa64(params.OrderID), q, nil)
}

// OrdersCancelAllRequest selects the scope whose orders to cancel.
type OrdersCancelAllRequest struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
	Stop      bool   `json:"stop"` // true cancels stop orders instead of regular ones
}

// CancelAll cancels every order for the given exchange/portfolio. Set
// Stop to true to cancel stop orders instead of regular ones. Returns only an
// error (bare success reply).
func (s *ordersService) CancelAll(
	ctx context.Context,
	params OrdersCancelAllRequest,
) error {
	q := restkit.NewValues().
		Str(keyExchange, &params.Exchange).
		Str(keyPortfolio, &params.Portfolio).
		Bool(keyStop, &params.Stop)
	return exec(ctx, s.c, http.MethodDelete, ordersBase+"/all", q, nil)
}
