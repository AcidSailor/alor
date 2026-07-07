package alor

import (
	"context"
	"net/http"

	"github.com/acidsailor/restkit"
)

// tradesService groups the Trades API operations. Obtain it via Client.Trades.
type tradesService struct{ c *Client }

// TradesGetRequest selects the exchange/portfolio and an optional REPO filter.
type TradesGetRequest struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
	WithRepo  *bool  `json:"withRepo,omitempty"` // include REPO trades
}

// Get returns today's trades for the given exchange/portfolio.
func (s *tradesService) Get(
	ctx context.Context,
	params TradesGetRequest,
) (*ResponseTradesV2Heavy, error) {
	q := heavyValues().Bool(keyWithRepo, params.WithRepo)
	return do[*ResponseTradesV2Heavy](ctx, s.c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/trades"), q, nil)
}

// TradesSymbolRequest selects the single-symbol trades and an optional
// instrument-group filter.
type TradesSymbolRequest struct {
	Exchange        string  `json:"exchange"`
	Portfolio       string  `json:"portfolio"`
	Symbol          string  `json:"symbol"`
	InstrumentGroup *string `json:"instrumentGroup,omitempty"` // board/instrument group, e.g. "TQBR"
}

// Symbol returns today's trades in a single symbol for the given
// exchange/portfolio.
func (s *tradesService) Symbol(
	ctx context.Context,
	params TradesSymbolRequest,
) (*ResponseTradesV2Heavy, error) {
	path := clientPath(
		params.Exchange,
		params.Portfolio,
		restkit.Pathf("/%s/trades", params.Symbol),
	)
	q := heavyValues().Str(keyInstrumentGroup, params.InstrumentGroup)
	return do[*ResponseTradesV2Heavy](ctx, s.c, http.MethodGet, path, q, nil)
}

// TradesHistoryRequest selects the exchange/portfolio and the optional
// trade-history filters.
type TradesHistoryRequest struct {
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

// History returns historical trades for the given exchange/portfolio.
func (s *tradesService) History(
	ctx context.Context,
	params TradesHistoryRequest,
) (*ResponseTradesV2Heavy, error) {
	path := restkit.Pathf(
		"/md/v2/Stats/%s/%s/history/trades",
		params.Exchange,
		params.Portfolio,
	)
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
	return do[*ResponseTradesV2Heavy](ctx, s.c, http.MethodGet, path, q, nil)
}

// TradesSymbolHistoryRequest selects the single-symbol trade history and the
// optional trade-history filters.
type TradesSymbolHistoryRequest struct {
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

// SymbolHistory returns historical trades in a single symbol for the
// given exchange/portfolio.
func (s *tradesService) SymbolHistory(
	ctx context.Context,
	params TradesSymbolHistoryRequest,
) (*ResponseTradesV2Heavy, error) {
	path := restkit.Pathf(
		"/md/v2/Stats/%s/%s/history/trades/%s",
		params.Exchange,
		params.Portfolio,
		params.Symbol,
	)
	q := heavyValues().
		Str(keyInstrumentGroup, params.InstrumentGroup).
		Int64(keyFrom, params.From).
		Int(keyLimit, params.Limit).
		Bool(keyOrderByTradeDate, params.OrderByTradeDate).
		Bool(keyDescending, params.Descending).
		Bool(keyWithRepo, params.WithRepo).
		Str(keySide, params.Side)
	q = setTime(q, keyDateFrom, params.DateFrom)
	return do[*ResponseTradesV2Heavy](ctx, s.c, http.MethodGet, path, q, nil)
}

// TradesAllRequest selects the instrument and the optional public-tape
// filters.
type TradesAllRequest struct {
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

// All returns the recent public trade tape for an instrument.
func (s *tradesService) All(
	ctx context.Context,
	params TradesAllRequest,
) (ResponseAllTradesHeavy, error) {
	path := restkit.Pathf(
		"/md/v2/Securities/%s/%s/alltrades",
		params.Exchange,
		params.Symbol,
	)
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
	return do[ResponseAllTradesHeavy](ctx, s.c, http.MethodGet, path, q, nil)
}

// TradesAllHistoryRequest selects the instrument, the required Limit cap, and
// the optional historical-tape filters.
type TradesAllHistoryRequest struct {
	Exchange        string  `json:"exchange"`
	Symbol          string  `json:"symbol"`
	Limit           int     `json:"limit"`                     // required cap on returned trades
	InstrumentGroup *string `json:"instrumentGroup,omitempty"` // board/instrument group, e.g. "TQBR"
	From            *int64  `json:"from,omitempty"`            // lower Unix-time (seconds) bound
	To              *int64  `json:"to,omitempty"`              // upper Unix-time (seconds) bound
	Offset          *int    `json:"offset,omitempty"`          // items to skip (pagination)
}

// AllHistory returns the historical public trade tape for an
// instrument; the required Limit caps the trade count.
func (s *tradesService) AllHistory(
	ctx context.Context,
	params TradesAllHistoryRequest,
) (*ResponseAllTradesHistoryHeavy, error) {
	path := restkit.Pathf(
		"/md/v2/Securities/%s/%s/alltrades/history",
		params.Exchange,
		params.Symbol,
	)
	q := heavyValues().
		Int(keyLimit, &params.Limit).
		Str(keyInstrumentGroup, params.InstrumentGroup).
		Int64(keyFrom, params.From).
		Int64(keyTo, params.To).
		Int(keyOffset, params.Offset)
	return do[*ResponseAllTradesHistoryHeavy](ctx, s.c, http.MethodGet,
		path, q, nil)
}
