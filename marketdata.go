package alor

import (
	"context"
	"net/http"
	"net/url"

	"github.com/acidsailor/restkit"
)

// marketDataService groups the MarketData API operations. Obtain it via Client.MarketData.
type marketDataService struct{ c *Client }

// MarketDataSearchRequest carries the optional free-text query.
type MarketDataSearchRequest struct {
	Query *string `json:"query,omitempty"` // free-text instrument filter; nil lists all
}

// Search returns instruments matching the free-text query; nil Query
// lists all.
func (s *marketDataService) Search(
	ctx context.Context,
	params MarketDataSearchRequest,
) (*ResponseSecuritiesHeavy, error) {
	q := heavyValues().Str(keyQuery, params.Query)
	return do[*ResponseSecuritiesHeavy](ctx, s.c, http.MethodGet,
		"/md/v2/Securities", q, nil)
}

// MarketDataSecuritiesByExchangeRequest selects the exchange and the optional listing
// filters.
type MarketDataSecuritiesByExchangeRequest struct {
	Exchange             string  `json:"exchange"`
	Market               *string `json:"market,omitempty"`               // market filter, e.g. "FORTS"
	IncludeOld           *bool   `json:"includeOld,omitempty"`           // include delisted/old instruments when true
	Limit                *int    `json:"limit,omitempty"`                // cap on returned items
	Offset               *int    `json:"offset,omitempty"`               // items to skip (pagination)
	IncludeNonBaseBoards *bool   `json:"includeNonBaseBoards,omitempty"` // include non-base trading boards when true
}

// SecuritiesByExchange lists instruments traded on the given exchange.
func (s *marketDataService) SecuritiesByExchange(
	ctx context.Context,
	params MarketDataSecuritiesByExchangeRequest,
) (*ResponseSecuritiesHeavy, error) {
	q := heavyValues().
		Str(keyMarket, params.Market).
		Bool(keyIncludeOld, params.IncludeOld).
		Int(keyLimit, params.Limit).
		Int(keyOffset, params.Offset).
		Bool(keyIncludeNonBaseBoards, params.IncludeNonBaseBoards)
	return do[*ResponseSecuritiesHeavy](ctx, s.c, http.MethodGet,
		"/md/v2/Securities/"+url.PathEscape(params.Exchange), q, nil)
}

// MarketDataSecurityRequest selects the single instrument and an optional
// instrument-group filter.
type MarketDataSecurityRequest struct {
	Exchange        string  `json:"exchange"`
	Symbol          string  `json:"symbol"`
	InstrumentGroup *string `json:"instrumentGroup,omitempty"` // board/instrument group, e.g. "TQBR"
}

// Security returns a single instrument on the given exchange.
func (s *marketDataService) Security(
	ctx context.Context,
	params MarketDataSecurityRequest,
) (*ResponseSecurityHeavy, error) {
	path := "/md/v2/Securities/" + url.PathEscape(params.Exchange) +
		"/" + url.PathEscape(params.Symbol)
	q := heavyValues().Str(keyInstrumentGroup, params.InstrumentGroup)
	return do[*ResponseSecurityHeavy](ctx, s.c, http.MethodGet, path, q, nil)
}

// MarketDataBoardsRequest selects the instrument whose boards to list.
type MarketDataBoardsRequest struct {
	Exchange string `json:"exchange"`
	Symbol   string `json:"symbol"`
}

// Boards returns the trading boards an instrument is available on.
func (s *marketDataService) Boards(
	ctx context.Context,
	params MarketDataBoardsRequest,
) (*ResponseAvailableBoards, error) {
	path := "/md/v2/Securities/" + url.PathEscape(params.Exchange) +
		"/" + url.PathEscape(params.Symbol) + "/availableBoards"
	return do[*ResponseAvailableBoards](ctx, s.c, http.MethodGet,
		path, restkit.NewValues(), nil)
}

// MarketDataFuturesQuoteRequest selects the futures instrument to quote.
type MarketDataFuturesQuoteRequest struct {
	Exchange string `json:"exchange"`
	Symbol   string `json:"symbol"`
}

// FuturesQuote returns the current quote for a futures instrument.
func (s *marketDataService) FuturesQuote(
	ctx context.Context,
	params MarketDataFuturesQuoteRequest,
) (*ResponseFuturesHeavy, error) {
	path := "/md/v2/Securities/" + url.PathEscape(params.Exchange) +
		"/" + url.PathEscape(params.Symbol) + "/actualFuturesQuote"
	return do[*ResponseFuturesHeavy](ctx, s.c, http.MethodGet,
		path, heavyValues(), nil)
}

// MarketDataQuotesRequest carries the comma-separated EXCHANGE:SYMBOL list.
type MarketDataQuotesRequest struct {
	Symbols string `json:"symbols"` // e.g. "MOEX:SBER,MOEX:GAZP"
}

// Quotes returns quotes for the instruments in Symbols, a comma-separated
// list of EXCHANGE:SYMBOL pairs (e.g. "MOEX:SBER,MOEX:GAZP").
func (s *marketDataService) Quotes(
	ctx context.Context,
	params MarketDataQuotesRequest,
) (*ResponseSymbolsHeavy, error) {
	path := "/md/v2/Securities/" + url.PathEscape(params.Symbols) + "/quotes"
	return do[*ResponseSymbolsHeavy](ctx, s.c, http.MethodGet,
		path, heavyValues(), nil)
}

// CurrencyPairs returns the tradable currency pairs.
func (s *marketDataService) CurrencyPairs(
	ctx context.Context,
) (*ResponseCurrencyPairs, error) {
	return do[*ResponseCurrencyPairs](ctx, s.c, http.MethodGet,
		"/md/v2/Securities/currencyPairs", restkit.NewValues(), nil)
}

// MarketDataOrderBookRequest selects the instrument and the optional instrument-group
// and depth filters.
type MarketDataOrderBookRequest struct {
	Exchange        string  `json:"exchange"`
	Symbol          string  `json:"symbol"`
	InstrumentGroup *string `json:"instrumentGroup,omitempty"` // board/instrument group, e.g. "TQBR"
	Depth           *int    `json:"depth,omitempty"`           // price levels per side
}

// OrderBook returns the order book (depth of market) for an instrument.
func (s *marketDataService) OrderBook(
	ctx context.Context,
	params MarketDataOrderBookRequest,
) (*ResponseOrderBookHeavy, error) {
	path := "/md/v2/orderbooks/" + url.PathEscape(params.Exchange) +
		"/" + url.PathEscape(params.Symbol)
	q := heavyValues().
		Str(keyInstrumentGroup, params.InstrumentGroup).
		Int(keyDepth, params.Depth)
	return do[*ResponseOrderBookHeavy](ctx, s.c, http.MethodGet, path, q, nil)
}

// MarketDataHistoryRequest selects the instrument and timeframe/window for the candle
// history. Exchange, Symbol, Tf, From, and To are required.
type MarketDataHistoryRequest struct {
	Exchange        string  `json:"exchange"`
	Symbol          string  `json:"symbol"`
	Tf              string  `json:"tf"`                        // timeframe, e.g. "60" (seconds) or "D"
	From            int64   `json:"from"`                      // lower Unix-time (seconds) bound
	To              int64   `json:"to"`                        // upper Unix-time (seconds) bound
	InstrumentGroup *string `json:"instrumentGroup,omitempty"` // board/instrument group, e.g. "TQBR"
	CountBack       *int    `json:"countBack,omitempty"`       // last N candles up to the To bound
	Untraded        *bool   `json:"untraded,omitempty"`        // include untraded candles when true
}

// History returns OHLCV candle history for an instrument. Exchange, Symbol,
// Tf (timeframe, e.g. "60" seconds or "D"), and the From/To Unix-time (seconds)
// window are required.
func (s *marketDataService) History(
	ctx context.Context,
	params MarketDataHistoryRequest,
) (*ResponseHistoryHeavy, error) {
	q := heavyValues().
		Str(keySymbol, &params.Symbol).
		Str(keyExchange, &params.Exchange).
		Str(keyTf, &params.Tf).
		Int64(keyFrom, &params.From).
		Int64(keyTo, &params.To).
		Str(keyInstrumentGroup, params.InstrumentGroup).
		Int(keyCountBack, params.CountBack).
		Bool(keyUntraded, params.Untraded)
	return do[*ResponseHistoryHeavy](ctx, s.c, http.MethodGet,
		"/md/v2/history", q, nil)
}
