package alor

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/acidsailor/restkit"
)

// The curated facade. Each method takes a single XxxParams value (required
// fields as values, optional query filters as pointers), delegates to the
// generic do transport pinning ?format=Heavy where the overlay collapsed the
// response oneOf to its _Heavy variant, and returns the generated response type
// directly. do surfaces a non-2xx as a *[ResponseError] and any other per-call
// failure as a *[RequestError] (cause preserved).

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

// GetOrdersParams selects the exchange/portfolio whose orders to return.
type GetOrdersParams struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
}

// GetOrders returns all orders for the given exchange/portfolio.
func (c *Client) GetOrders(
	ctx context.Context,
	params GetOrdersParams,
) (ResponseOrdersHeavy, error) {
	return do[ResponseOrdersHeavy](ctx, c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/orders"),
		heavyValues(), nil)
}

// GetOrderParams selects the single order to return.
type GetOrderParams struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
	OrderID   int64  `json:"orderId"`
}

// GetOrder returns a single order by id for the given exchange/portfolio.
func (c *Client) GetOrder(
	ctx context.Context,
	params GetOrderParams,
) (ResponseOrderHeavy, error) {
	path := clientPath(params.Exchange, params.Portfolio,
		"/orders/"+strconv.FormatInt(params.OrderID, 10))
	return do[ResponseOrderHeavy](ctx, c, http.MethodGet,
		path, heavyValues(), nil)
}

// GetSummaryParams selects the exchange/portfolio to summarize.
type GetSummaryParams struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
}

// GetSummary returns the portfolio summary for the given exchange/portfolio.
func (c *Client) GetSummary(
	ctx context.Context,
	params GetSummaryParams,
) (ResponseSummaryHeavy, error) {
	return do[ResponseSummaryHeavy](ctx, c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/summary"),
		heavyValues(), nil)
}

// GetPositionsParams selects the exchange/portfolio whose positions to return.
type GetPositionsParams struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
}

// GetPositions returns all positions for the given exchange/portfolio.
func (c *Client) GetPositions(
	ctx context.Context,
	params GetPositionsParams,
) (ResponsePositionsHeavy, error) {
	return do[ResponsePositionsHeavy](ctx, c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/positions"),
		heavyValues(), nil)
}

// GetRiskParams selects the exchange/portfolio whose risk metrics to return.
type GetRiskParams struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
}

// GetRisk returns risk metrics for the given exchange/portfolio.
func (c *Client) GetRisk(
	ctx context.Context,
	params GetRiskParams,
) (ResponseRiskHeavy, error) {
	return do[ResponseRiskHeavy](ctx, c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/risk"),
		heavyValues(), nil)
}

// GetStopOrdersParams selects the exchange/portfolio whose stop orders to
// return.
type GetStopOrdersParams struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
}

// GetStopOrders returns all stop orders for the given exchange/portfolio.
func (c *Client) GetStopOrders(
	ctx context.Context,
	params GetStopOrdersParams,
) (ResponseStopOrdersWarpHeavy, error) {
	return do[ResponseStopOrdersWarpHeavy](ctx, c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/stoporders"),
		heavyValues(), nil)
}

// GetStopOrderParams selects the single stop order to return.
type GetStopOrderParams struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
	OrderID   int64  `json:"orderId"`
}

// GetStopOrder returns a single stop order by id for the given
// exchange/portfolio.
func (c *Client) GetStopOrder(
	ctx context.Context,
	params GetStopOrderParams,
) (ResponseStopOrderWarpHeavy, error) {
	path := clientPath(params.Exchange, params.Portfolio,
		"/stoporders/"+strconv.FormatInt(params.OrderID, 10))
	return do[ResponseStopOrderWarpHeavy](ctx, c, http.MethodGet,
		path, heavyValues(), nil)
}

// SearchSecuritiesParams carries the optional free-text query.
type SearchSecuritiesParams struct {
	Query *string `json:"query,omitempty"` // free-text instrument filter; nil lists all
}

// SearchSecurities returns instruments matching the free-text query; nil Query
// lists all.
func (c *Client) SearchSecurities(
	ctx context.Context,
	params SearchSecuritiesParams,
) (ResponseSecuritiesHeavy, error) {
	q := heavyValues().Str(keyQuery, params.Query)
	return do[ResponseSecuritiesHeavy](ctx, c, http.MethodGet,
		"/md/v2/Securities", q, nil)
}

// ListOrderGroups returns all order groups (linked-order baskets).
func (c *Client) ListOrderGroups(
	ctx context.Context,
) ([]ResponseOrderGroupInfo, error) {
	return do[[]ResponseOrderGroupInfo](ctx, c, http.MethodGet,
		"/commandapi/api/orderGroups", restkit.NewValues(), nil)
}

// GetOrderGroupParams selects the single order group to return.
type GetOrderGroupParams struct {
	OrderGroupID string `json:"orderGroupId"`
}

// GetOrderGroup returns a single order group by id.
func (c *Client) GetOrderGroup(
	ctx context.Context,
	params GetOrderGroupParams,
) (ResponseOrderGroupInfo, error) {
	return do[ResponseOrderGroupInfo](ctx, c, http.MethodGet,
		"/commandapi/api/orderGroups/"+url.PathEscape(params.OrderGroupID),
		restkit.NewValues(), nil)
}

// GetPositionParams selects the single-symbol position to return.
type GetPositionParams struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
	Symbol    string `json:"symbol"`
}

// GetPosition returns the position in a single symbol for the given
// exchange/portfolio.
func (c *Client) GetPosition(
	ctx context.Context,
	params GetPositionParams,
) (ResponsePositionHeavy, error) {
	path := clientPath(
		params.Exchange,
		params.Portfolio,
		"/positions/"+url.PathEscape(params.Symbol),
	)
	return do[ResponsePositionHeavy](ctx, c, http.MethodGet,
		path, heavyValues(), nil)
}

// GetPositionsByLoginParams selects the trade-account login and an optional
// currency filter.
type GetPositionsByLoginParams struct {
	Login           string `json:"login"`
	WithoutCurrency *bool  `json:"withoutCurrency,omitempty"` // drop currency positions when true
}

// GetPositionsByLogin returns all positions for the given trade-account login,
// across its portfolios.
func (c *Client) GetPositionsByLogin(
	ctx context.Context,
	params GetPositionsByLoginParams,
) (ResponsePositionsHeavy, error) {
	path := "/md/v2/Clients/" + url.PathEscape(params.Login) + "/positions"
	q := heavyValues().Bool(keyWithoutCurrency, params.WithoutCurrency)
	return do[ResponsePositionsHeavy](ctx, c, http.MethodGet, path, q, nil)
}

// GetFortsRiskParams selects the exchange/portfolio whose FORTS risk to return.
type GetFortsRiskParams struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
}

// GetFortsRisk returns FORTS (derivatives-market) risk metrics for the given
// exchange/portfolio.
func (c *Client) GetFortsRisk(
	ctx context.Context,
	params GetFortsRiskParams,
) (ResponseFortsRiskHeavy, error) {
	return do[ResponseFortsRiskHeavy](ctx, c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/fortsrisk"),
		heavyValues(), nil)
}

// GetTradesParams selects the exchange/portfolio and an optional REPO filter.
type GetTradesParams struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
	WithRepo  *bool  `json:"withRepo,omitempty"` // include REPO trades
}

// GetTrades returns today's trades for the given exchange/portfolio.
func (c *Client) GetTrades(
	ctx context.Context,
	params GetTradesParams,
) (ResponseTradesV2Heavy, error) {
	q := heavyValues().Bool(keyWithRepo, params.WithRepo)
	return do[ResponseTradesV2Heavy](ctx, c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/trades"), q, nil)
}

// GetSymbolTradesParams selects the single-symbol trades and an optional
// instrument-group filter.
type GetSymbolTradesParams struct {
	Exchange        string  `json:"exchange"`
	Portfolio       string  `json:"portfolio"`
	Symbol          string  `json:"symbol"`
	InstrumentGroup *string `json:"instrumentGroup,omitempty"` // board/instrument group, e.g. "TQBR"
}

// GetSymbolTrades returns today's trades in a single symbol for the given
// exchange/portfolio.
func (c *Client) GetSymbolTrades(
	ctx context.Context,
	params GetSymbolTradesParams,
) (ResponseTradesV2Heavy, error) {
	path := clientPath(
		params.Exchange,
		params.Portfolio,
		"/"+url.PathEscape(params.Symbol)+"/trades",
	)
	q := heavyValues().Str(keyInstrumentGroup, params.InstrumentGroup)
	return do[ResponseTradesV2Heavy](ctx, c, http.MethodGet, path, q, nil)
}

// GetTradeHistoryParams selects the exchange/portfolio and the optional
// trade-history filters.
type GetTradeHistoryParams struct {
	Exchange         string  `json:"exchange"`
	Portfolio        string  `json:"portfolio"`
	InstrumentGroup  *string `json:"instrumentGroup,omitempty"`  // board/instrument group, e.g. "TQBR"
	DateFrom         *Time   `json:"dateFrom,omitempty"`         // start instant, sent as RFC3339 UTC; zero/nil omits
	Ticker           *string `json:"ticker,omitempty"`           // ticker symbol filter
	From             *int64  `json:"from,omitempty"`             // lower Unix-time (seconds) bound
	Limit            *int    `json:"limit,omitempty"`            // cap on returned items
	OrderByTradeDate *bool   `json:"orderByTradeDate,omitempty"` // order by trade date when true
	Descending       *bool   `json:"descending,omitempty"`       // reverse sort order when true
	WithRepo         *bool   `json:"withRepo,omitempty"`         // include REPO trades when true
	Side             *string `json:"side,omitempty"`             // trade direction, "buy" or "sell"
}

// GetTradeHistory returns historical trades for the given exchange/portfolio.
func (c *Client) GetTradeHistory(
	ctx context.Context,
	params GetTradeHistoryParams,
) (ResponseTradesV2Heavy, error) {
	path := "/md/v2/Stats/" + url.PathEscape(params.Exchange) +
		"/" + url.PathEscape(params.Portfolio) + "/history/trades"
	q := heavyValues().
		Str(keyInstrumentGroup, params.InstrumentGroup).
		Str(keyTicker, params.Ticker).
		Int64(keyFrom, params.From).
		Int(keyLimit, params.Limit).
		Bool(keyOrderByTradeDate, params.OrderByTradeDate).
		Bool(keyDescending, params.Descending).
		Bool(keyWithRepo, params.WithRepo).
		Str(keySide, params.Side)
	q = setTime(q, keyDateFrom, params.DateFrom)
	return do[ResponseTradesV2Heavy](ctx, c, http.MethodGet, path, q, nil)
}

// GetSymbolTradeHistoryParams selects the single-symbol trade history and the
// optional trade-history filters.
type GetSymbolTradeHistoryParams struct {
	Exchange         string  `json:"exchange"`
	Portfolio        string  `json:"portfolio"`
	Symbol           string  `json:"symbol"`
	InstrumentGroup  *string `json:"instrumentGroup,omitempty"`  // board/instrument group, e.g. "TQBR"
	DateFrom         *Time   `json:"dateFrom,omitempty"`         // start instant, sent as RFC3339 UTC; zero/nil omits
	From             *int64  `json:"from,omitempty"`             // lower Unix-time (seconds) bound
	Limit            *int    `json:"limit,omitempty"`            // cap on returned items
	OrderByTradeDate *bool   `json:"orderByTradeDate,omitempty"` // order by trade date when true
	Descending       *bool   `json:"descending,omitempty"`       // reverse sort order when true
	WithRepo         *bool   `json:"withRepo,omitempty"`         // include REPO trades when true
	Side             *string `json:"side,omitempty"`             // trade direction, "buy" or "sell"
}

// GetSymbolTradeHistory returns historical trades in a single symbol for the
// given exchange/portfolio.
func (c *Client) GetSymbolTradeHistory(
	ctx context.Context,
	params GetSymbolTradeHistoryParams,
) (ResponseTradesV2Heavy, error) {
	path := "/md/v2/Stats/" + url.PathEscape(params.Exchange) +
		"/" + url.PathEscape(params.Portfolio) + "/history/trades/" +
		url.PathEscape(params.Symbol)
	q := heavyValues().
		Str(keyInstrumentGroup, params.InstrumentGroup).
		Int64(keyFrom, params.From).
		Int(keyLimit, params.Limit).
		Bool(keyOrderByTradeDate, params.OrderByTradeDate).
		Bool(keyDescending, params.Descending).
		Bool(keyWithRepo, params.WithRepo).
		Str(keySide, params.Side)
	q = setTime(q, keyDateFrom, params.DateFrom)
	return do[ResponseTradesV2Heavy](ctx, c, http.MethodGet, path, q, nil)
}

// GetSecuritiesByExchangeParams selects the exchange and the optional listing
// filters.
type GetSecuritiesByExchangeParams struct {
	Exchange             string  `json:"exchange"`
	Market               *string `json:"market,omitempty"`               // market filter, e.g. "FORTS"
	IncludeOld           *bool   `json:"includeOld,omitempty"`           // include delisted/old instruments when true
	Limit                *int    `json:"limit,omitempty"`                // cap on returned items
	Offset               *int    `json:"offset,omitempty"`               // items to skip (pagination)
	IncludeNonBaseBoards *bool   `json:"includeNonBaseBoards,omitempty"` // include non-base trading boards when true
}

// GetSecuritiesByExchange lists instruments traded on the given exchange.
func (c *Client) GetSecuritiesByExchange(
	ctx context.Context,
	params GetSecuritiesByExchangeParams,
) (ResponseSecuritiesHeavy, error) {
	q := heavyValues().
		Str(keyMarket, params.Market).
		Bool(keyIncludeOld, params.IncludeOld).
		Int(keyLimit, params.Limit).
		Int(keyOffset, params.Offset).
		Bool(keyIncludeNonBaseBoards, params.IncludeNonBaseBoards)
	return do[ResponseSecuritiesHeavy](ctx, c, http.MethodGet,
		"/md/v2/Securities/"+url.PathEscape(params.Exchange), q, nil)
}

// GetSecurityParams selects the single instrument and an optional
// instrument-group filter.
type GetSecurityParams struct {
	Exchange        string  `json:"exchange"`
	Symbol          string  `json:"symbol"`
	InstrumentGroup *string `json:"instrumentGroup,omitempty"` // board/instrument group, e.g. "TQBR"
}

// GetSecurity returns a single instrument on the given exchange.
func (c *Client) GetSecurity(
	ctx context.Context,
	params GetSecurityParams,
) (ResponseSecurityHeavy, error) {
	path := "/md/v2/Securities/" + url.PathEscape(params.Exchange) +
		"/" + url.PathEscape(params.Symbol)
	q := heavyValues().Str(keyInstrumentGroup, params.InstrumentGroup)
	return do[ResponseSecurityHeavy](ctx, c, http.MethodGet, path, q, nil)
}

// GetAvailableBoardsParams selects the instrument whose boards to list.
type GetAvailableBoardsParams struct {
	Exchange string `json:"exchange"`
	Symbol   string `json:"symbol"`
}

// GetAvailableBoards returns the trading boards an instrument is available on.
func (c *Client) GetAvailableBoards(
	ctx context.Context,
	params GetAvailableBoardsParams,
) (ResponseAvailableBoards, error) {
	path := "/md/v2/Securities/" + url.PathEscape(params.Exchange) +
		"/" + url.PathEscape(params.Symbol) + "/availableBoards"
	return do[ResponseAvailableBoards](ctx, c, http.MethodGet,
		path, restkit.NewValues(), nil)
}

// GetAllTradesParams selects the instrument and the optional public-tape
// filters.
type GetAllTradesParams struct {
	Exchange             string   `json:"exchange"`
	Symbol               string   `json:"symbol"`
	InstrumentGroup      *string  `json:"instrumentGroup,omitempty"`      // board/instrument group, e.g. "TQBR"
	From                 *int64   `json:"from,omitempty"`                 // lower Unix-time (seconds) bound
	To                   *int64   `json:"to,omitempty"`                   // upper Unix-time (seconds) bound
	FromID               *int64   `json:"fromId,omitempty"`               // lower trade-id bound (inclusive)
	ToID                 *int64   `json:"toId,omitempty"`                 // upper trade-id bound (inclusive)
	QtyFrom              *int64   `json:"qtyFrom,omitempty"`              // minimum trade quantity
	QtyTo                *int64   `json:"qtyTo,omitempty"`                // maximum trade quantity
	PriceFrom            *float64 `json:"priceFrom,omitempty"`            // minimum trade price
	PriceTo              *float64 `json:"priceTo,omitempty"`              // maximum trade price
	Side                 *string  `json:"side,omitempty"`                 // trade direction, "buy" or "sell"
	Offset               *int     `json:"offset,omitempty"`               // items to skip (pagination)
	Take                 *int     `json:"take,omitempty"`                 // cap on returned items
	Descending           *bool    `json:"descending,omitempty"`           // reverse sort order when true
	IncludeVirtualTrades *bool    `json:"includeVirtualTrades,omitempty"` // include virtual (indicative) trades when true
}

// GetAllTrades returns the recent public trade tape for an instrument.
func (c *Client) GetAllTrades(
	ctx context.Context,
	params GetAllTradesParams,
) (ResponseAllTradesHeavy, error) {
	path := "/md/v2/Securities/" + url.PathEscape(params.Exchange) +
		"/" + url.PathEscape(params.Symbol) + "/alltrades"
	q := heavyValues().
		Str(keyInstrumentGroup, params.InstrumentGroup).
		Int64(keyFrom, params.From).
		Int64(keyTo, params.To).
		Int64(keyFromID, params.FromID).
		Int64(keyToID, params.ToID).
		Int64(keyQtyFrom, params.QtyFrom).
		Int64(keyQtyTo, params.QtyTo).
		Float(keyPriceFrom, params.PriceFrom).
		Float(keyPriceTo, params.PriceTo).
		Str(keySide, params.Side).
		Int(keyOffset, params.Offset).
		Int(keyTake, params.Take).
		Bool(keyDescending, params.Descending).
		Bool(keyIncludeVirtualTrades, params.IncludeVirtualTrades)
	return do[ResponseAllTradesHeavy](ctx, c, http.MethodGet, path, q, nil)
}

// GetAllTradesHistoryParams selects the instrument, the required Limit cap, and
// the optional historical-tape filters.
type GetAllTradesHistoryParams struct {
	Exchange        string  `json:"exchange"`
	Symbol          string  `json:"symbol"`
	Limit           int     `json:"limit"`                     // required cap on returned trades
	InstrumentGroup *string `json:"instrumentGroup,omitempty"` // board/instrument group, e.g. "TQBR"
	From            *int64  `json:"from,omitempty"`            // lower Unix-time (seconds) bound
	To              *int64  `json:"to,omitempty"`              // upper Unix-time (seconds) bound
	Offset          *int    `json:"offset,omitempty"`          // items to skip (pagination)
}

// GetAllTradesHistory returns the historical public trade tape for an
// instrument; the required Limit caps the trade count.
func (c *Client) GetAllTradesHistory(
	ctx context.Context,
	params GetAllTradesHistoryParams,
) (ResponseAllTradesHistoryHeavy, error) {
	path := "/md/v2/Securities/" + url.PathEscape(params.Exchange) +
		"/" + url.PathEscape(params.Symbol) + "/alltrades/history"
	q := heavyValues().
		Int(keyLimit, &params.Limit).
		Str(keyInstrumentGroup, params.InstrumentGroup).
		Int64(keyFrom, params.From).
		Int64(keyTo, params.To).
		Int(keyOffset, params.Offset)
	return do[ResponseAllTradesHistoryHeavy](ctx, c, http.MethodGet,
		path, q, nil)
}

// GetActualFuturesQuoteParams selects the futures instrument to quote.
type GetActualFuturesQuoteParams struct {
	Exchange string `json:"exchange"`
	Symbol   string `json:"symbol"`
}

// GetActualFuturesQuote returns the current quote for a futures instrument.
func (c *Client) GetActualFuturesQuote(
	ctx context.Context,
	params GetActualFuturesQuoteParams,
) (ResponseFuturesHeavy, error) {
	path := "/md/v2/Securities/" + url.PathEscape(params.Exchange) +
		"/" + url.PathEscape(params.Symbol) + "/actualFuturesQuote"
	return do[ResponseFuturesHeavy](ctx, c, http.MethodGet,
		path, heavyValues(), nil)
}

// GetQuotesParams carries the comma-separated EXCHANGE:SYMBOL list.
type GetQuotesParams struct {
	Symbols string `json:"symbols"` // e.g. "MOEX:SBER,MOEX:GAZP"
}

// GetQuotes returns quotes for the instruments in Symbols, a comma-separated
// list of EXCHANGE:SYMBOL pairs (e.g. "MOEX:SBER,MOEX:GAZP").
func (c *Client) GetQuotes(
	ctx context.Context,
	params GetQuotesParams,
) (ResponseSymbolsHeavy, error) {
	path := "/md/v2/Securities/" + url.PathEscape(params.Symbols) + "/quotes"
	return do[ResponseSymbolsHeavy](ctx, c, http.MethodGet,
		path, heavyValues(), nil)
}

// GetCurrencyPairs returns the tradable currency pairs.
func (c *Client) GetCurrencyPairs(
	ctx context.Context,
) (ResponseCurrencyPairs, error) {
	return do[ResponseCurrencyPairs](ctx, c, http.MethodGet,
		"/md/v2/Securities/currencyPairs", restkit.NewValues(), nil)
}

// GetOrderBookParams selects the instrument and the optional instrument-group
// and depth filters.
type GetOrderBookParams struct {
	Exchange        string  `json:"exchange"`
	Symbol          string  `json:"symbol"`
	InstrumentGroup *string `json:"instrumentGroup,omitempty"` // board/instrument group, e.g. "TQBR"
	Depth           *int    `json:"depth,omitempty"`           // price levels per side
}

// GetOrderBook returns the order book (depth of market) for an instrument.
func (c *Client) GetOrderBook(
	ctx context.Context,
	params GetOrderBookParams,
) (ResponseOrderBookHeavy, error) {
	path := "/md/v2/orderbooks/" + url.PathEscape(params.Exchange) +
		"/" + url.PathEscape(params.Symbol)
	q := heavyValues().
		Str(keyInstrumentGroup, params.InstrumentGroup).
		Int(keyDepth, params.Depth)
	return do[ResponseOrderBookHeavy](ctx, c, http.MethodGet, path, q, nil)
}

// GetRiskRatesParams selects the required risk category and the optional
// risk-rate filters.
type GetRiskRatesParams struct {
	RiskCategoryID int     `json:"riskCategoryId"`     // required risk category id
	Exchange       *string `json:"exchange,omitempty"` // exchange-code filter
	Ticker         *string `json:"ticker,omitempty"`   // ticker symbol filter
	Search         *string `json:"search,omitempty"`   // free-text search filter
	Limit          *int    `json:"limit,omitempty"`    // cap on returned items
	Offset         *int    `json:"offset,omitempty"`   // items to skip (pagination)
}

// GetRiskRates returns risk rates for the given risk category. RiskCategoryID
// is required.
func (c *Client) GetRiskRates(
	ctx context.Context,
	params GetRiskRatesParams,
) (ResponseRiskRates, error) {
	q := restkit.NewValues().
		Int(keyRiskCategoryID, &params.RiskCategoryID).
		Str(keyExchange, params.Exchange).
		Str(keyTicker, params.Ticker).
		Str(keySearch, params.Search).
		Int(keyLimit, params.Limit).
		Int(keyOffset, params.Offset)
	return do[ResponseRiskRates](ctx, c, http.MethodGet,
		"/md/v2/risk/rates", q, nil)
}

// GetHistoryParams selects the instrument and timeframe/window for the candle
// history. Exchange, Symbol, Tf, From, and To are required.
type GetHistoryParams struct {
	Exchange        string  `json:"exchange"`
	Symbol          string  `json:"symbol"`
	Tf              string  `json:"tf"`                        // timeframe, e.g. "60" (seconds) or "D"
	From            int64   `json:"from"`                      // lower Unix-time (seconds) bound
	To              int64   `json:"to"`                        // upper Unix-time (seconds) bound
	InstrumentGroup *string `json:"instrumentGroup,omitempty"` // board/instrument group, e.g. "TQBR"
	CountBack       *int    `json:"countBack,omitempty"`       // last N candles up to the To bound
	Untraded        *bool   `json:"untraded,omitempty"`        // include untraded candles when true
}

// GetHistory returns OHLCV candle history for an instrument. Exchange, Symbol,
// Tf (timeframe, e.g. "60" seconds or "D"), and the From/To Unix-time (seconds)
// window are required.
func (c *Client) GetHistory(
	ctx context.Context,
	params GetHistoryParams,
) (ResponseHistoryHeavy, error) {
	q := heavyValues().
		Str(keySymbol, &params.Symbol).
		Str(keyExchange, &params.Exchange).
		Str(keyTf, &params.Tf).
		Int64(keyFrom, &params.From).
		Int64(keyTo, &params.To).
		Str(keyInstrumentGroup, params.InstrumentGroup).
		Int(keyCountBack, params.CountBack).
		Bool(keyUntraded, params.Untraded)
	return do[ResponseHistoryHeavy](ctx, c, http.MethodGet,
		"/md/v2/history", q, nil)
}
