package alor

import (
	"context"
	"net/http"
	"net/url"

	"github.com/acidsailor/restkit"
)

// The mutating facade: order place/replace/cancel, cost estimation, and
// order-group management. Each method takes a single XxxParams value and
// delegates to the generic transport — [do] for JSON payloads, [exec] for the
// bare text/plain "success" replies — so a non-2xx surfaces as a
// *[ResponseError] and any other per-call failure as a *[RequestError].
//
// Order-command endpoints (place/replace) require an idempotency key sent as the
// X-REQID header: use a unique ReqID per logical command and reuse it on retries
// so a resend cannot double-submit. Bodies and responses are the generated
// structs under their generated names.

// ordersBase is the common prefix for the CommandAPI order-action endpoints.
const ordersBase = "/commandapi/warptrans/TRADE/v2/client/orders"

// PlaceMarketOrderParams carries the idempotency key and the market-order body.
type PlaceMarketOrderParams struct {
	ReqID string                `json:"-"`     // X-REQID idempotency key, caller-minted
	Order OrdersActionsMarketTV `json:"order"` // request body
}

// PlaceMarketOrder submits a new market order.
func (c *Client) PlaceMarketOrder(
	ctx context.Context,
	params PlaceMarketOrderParams,
) (ResponseOrderActionLimitMarketCommandAPI, error) {
	return do[ResponseOrderActionLimitMarketCommandAPI](ctx, c, http.MethodPost,
		ordersBase+"/actions/market", restkit.NewValues(), params.Order,
		withReqID(params.ReqID))
}

// PlaceLimitOrderParams carries the idempotency key and the limit-order body.
type PlaceLimitOrderParams struct {
	ReqID string                   `json:"-"`
	Order OrdersActionsLimitTVPost `json:"order"`
}

// PlaceLimitOrder submits a new limit order.
func (c *Client) PlaceLimitOrder(
	ctx context.Context,
	params PlaceLimitOrderParams,
) (ResponseOrderActionLimitMarketCommandAPI, error) {
	return do[ResponseOrderActionLimitMarketCommandAPI](ctx, c, http.MethodPost,
		ordersBase+"/actions/limit", restkit.NewValues(), params.Order,
		withReqID(params.ReqID))
}

// PlaceStopOrderParams carries the idempotency key and the stop-order body.
type PlaceStopOrderParams struct {
	ReqID string                        `json:"-"`
	Order OrdersActionsStopMarketTVWarp `json:"order"`
}

// PlaceStopOrder submits a new stop (stop-market) order.
func (c *Client) PlaceStopOrder(
	ctx context.Context,
	params PlaceStopOrderParams,
) (ResponseOrderActionLimitMarketCommandAPI, error) {
	return do[ResponseOrderActionLimitMarketCommandAPI](ctx, c, http.MethodPost,
		ordersBase+"/actions/stop", restkit.NewValues(), params.Order,
		withReqID(params.ReqID))
}

// PlaceStopLimitOrderParams carries the idempotency key and the stop-limit-order
// body.
type PlaceStopLimitOrderParams struct {
	ReqID string                           `json:"-"`
	Order OrdersActionsStopLimitTVWarpPost `json:"order"`
}

// PlaceStopLimitOrder submits a new stop-limit order.
func (c *Client) PlaceStopLimitOrder(
	ctx context.Context,
	params PlaceStopLimitOrderParams,
) (ResponseOrderActionLimitMarketCommandAPI, error) {
	return do[ResponseOrderActionLimitMarketCommandAPI](ctx, c, http.MethodPost,
		ordersBase+"/actions/stopLimit", restkit.NewValues(), params.Order,
		withReqID(params.ReqID))
}

// ReplaceMarketOrderParams carries the idempotency key, the target order id, and
// the replacement market-order body.
type ReplaceMarketOrderParams struct {
	ReqID   string                `json:"-"`
	OrderID int64                 `json:"orderId"` // path
	Order   OrdersActionsMarketTV `json:"order"`
}

// ReplaceMarketOrder replaces the market order with the given id.
func (c *Client) ReplaceMarketOrder(
	ctx context.Context,
	params ReplaceMarketOrderParams,
) (ResponseOrderActionLimitMarket, error) {
	path := ordersBase + "/actions/market/" + itoa64(params.OrderID)
	return do[ResponseOrderActionLimitMarket](ctx, c, http.MethodPut,
		path, restkit.NewValues(), params.Order, withReqID(params.ReqID))
}

// ReplaceLimitOrderParams carries the idempotency key, the target order id, and
// the replacement limit-order body.
type ReplaceLimitOrderParams struct {
	ReqID   string                  `json:"-"`
	OrderID int64                   `json:"orderId"` // path
	Order   OrdersActionsLimitTVPut `json:"order"`
}

// ReplaceLimitOrder replaces the limit order with the given id.
func (c *Client) ReplaceLimitOrder(
	ctx context.Context,
	params ReplaceLimitOrderParams,
) (ResponseOrderActionLimitMarket, error) {
	path := ordersBase + "/actions/limit/" + itoa64(params.OrderID)
	return do[ResponseOrderActionLimitMarket](ctx, c, http.MethodPut,
		path, restkit.NewValues(), params.Order, withReqID(params.ReqID))
}

// ReplaceStopOrderParams carries the idempotency key, the target stop-order id,
// and the replacement stop-order body.
type ReplaceStopOrderParams struct {
	ReqID       string                        `json:"-"`
	StopOrderID int64                         `json:"stopOrderId"` // path
	Order       OrdersActionsStopMarketTVWarp `json:"order"`
}

// ReplaceStopOrder replaces the stop order with the given id.
func (c *Client) ReplaceStopOrder(
	ctx context.Context,
	params ReplaceStopOrderParams,
) (ResponseOrderActionLimitMarketCommandAPI, error) {
	path := ordersBase + "/actions/stop/" + itoa64(params.StopOrderID)
	return do[ResponseOrderActionLimitMarketCommandAPI](ctx, c, http.MethodPut,
		path, restkit.NewValues(), params.Order, withReqID(params.ReqID))
}

// ReplaceStopLimitOrderParams carries the idempotency key, the target
// stop-order id, and the replacement stop-limit-order body.
type ReplaceStopLimitOrderParams struct {
	ReqID       string                          `json:"-"`
	StopOrderID int64                           `json:"stopOrderId"` // path
	Order       OrdersActionsStopLimitTVWarpPut `json:"order"`
}

// ReplaceStopLimitOrder replaces the stop-limit order with the given id.
func (c *Client) ReplaceStopLimitOrder(
	ctx context.Context,
	params ReplaceStopLimitOrderParams,
) (ResponseOrderActionLimitMarketCommandAPI, error) {
	path := ordersBase + "/actions/stopLimit/" + itoa64(params.StopOrderID)
	return do[ResponseOrderActionLimitMarketCommandAPI](ctx, c, http.MethodPut,
		path, restkit.NewValues(), params.Order, withReqID(params.ReqID))
}

// EstimateOrderParams carries the order to estimate.
type EstimateOrderParams struct {
	Order EstimateOrder `json:"order"`
}

// EstimateOrder estimates the cost and lot count of a prospective order without
// submitting it.
func (c *Client) EstimateOrder(
	ctx context.Context,
	params EstimateOrderParams,
) (ResponseEstimateOrder, error) {
	return do[ResponseEstimateOrder](ctx, c, http.MethodPost,
		ordersBase+"/estimate", restkit.NewValues(), params.Order)
}

// EstimateOrdersParams carries the orders to estimate.
type EstimateOrdersParams struct {
	Orders EstimateOrders `json:"orders"`
}

// EstimateOrders estimates the cost and lot count of several prospective orders
// in one call.
func (c *Client) EstimateOrders(
	ctx context.Context,
	params EstimateOrdersParams,
) (ResponseEstimateOrders, error) {
	return do[ResponseEstimateOrders](ctx, c, http.MethodPost,
		ordersBase+"/estimate/all", restkit.NewValues(), params.Orders)
}

// CancelOrderParams selects the order to cancel and its scope.
type CancelOrderParams struct {
	OrderID   int64  `json:"orderId"`   // path
	Exchange  string `json:"exchange"`  // query
	Portfolio string `json:"portfolio"` // query
	Stop      bool   `json:"stop"`      // query; true cancels a stop order
}

// CancelOrder cancels the order with the given id. Exchange and Portfolio scope
// it; set Stop to true for a stop order. Returns only an error (bare success
// reply).
func (c *Client) CancelOrder(
	ctx context.Context,
	params CancelOrderParams,
) error {
	q := restkit.NewValues().
		Str(keyExchange, &params.Exchange).
		Str(keyPortfolio, &params.Portfolio).
		Bool(keyStop, &params.Stop)
	return exec(ctx, c, http.MethodDelete,
		ordersBase+"/"+itoa64(params.OrderID), q, nil)
}

// CancelAllOrdersParams selects the scope whose orders to cancel.
type CancelAllOrdersParams struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
	Stop      bool   `json:"stop"` // true cancels stop orders instead of regular ones
}

// CancelAllOrders cancels every order for the given exchange/portfolio. Set
// Stop to true to cancel stop orders instead of regular ones. Returns only an
// error (bare success reply).
func (c *Client) CancelAllOrders(
	ctx context.Context,
	params CancelAllOrdersParams,
) error {
	q := restkit.NewValues().
		Str(keyExchange, &params.Exchange).
		Str(keyPortfolio, &params.Portfolio).
		Bool(keyStop, &params.Stop)
	return exec(ctx, c, http.MethodDelete, ordersBase+"/all", q, nil)
}

// CreateOrderGroupParams carries the order group to create.
type CreateOrderGroupParams struct {
	Group OrderGroupCreate `json:"group"`
}

// CreateOrderGroup links the given orders into a new order group and returns its
// id.
func (c *Client) CreateOrderGroup(
	ctx context.Context,
	params CreateOrderGroupParams,
) (ResponseOrderGroupCreationSuccess, error) {
	return do[ResponseOrderGroupCreationSuccess](ctx, c, http.MethodPost,
		"/commandapi/api/orderGroups", restkit.NewValues(), params.Group)
}

// UpdateOrderGroupParams selects the order group to modify and carries the
// modification.
type UpdateOrderGroupParams struct {
	OrderGroupID string           `json:"orderGroupId"` // path
	Group        OrderGroupModify `json:"group"`
}

// UpdateOrderGroup modifies the order group with the given id. Returns only an
// error (bare success reply).
func (c *Client) UpdateOrderGroup(
	ctx context.Context,
	params UpdateOrderGroupParams,
) error {
	path := "/commandapi/api/orderGroups/" + url.PathEscape(params.OrderGroupID)
	return exec(ctx, c, http.MethodPut, path, restkit.NewValues(), params.Group)
}

// DeleteOrderGroupParams selects the order group to delete.
type DeleteOrderGroupParams struct {
	OrderGroupID string `json:"orderGroupId"` // path
}

// DeleteOrderGroup deletes the order group with the given id. Returns only an
// error (bare success reply).
func (c *Client) DeleteOrderGroup(
	ctx context.Context,
	params DeleteOrderGroupParams,
) error {
	path := "/commandapi/api/orderGroups/" + url.PathEscape(params.OrderGroupID)
	return exec(ctx, c, http.MethodDelete, path, restkit.NewValues(), nil)
}
