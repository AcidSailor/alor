package alor

import (
	"strconv"

	"github.com/acidsailor/restkit"
)

// Query-parameter key constants. Separate from struct tags because the facade
// methods reference them when building each request's restkit.Values.
const (
	keyFormat = "format"
	keyQuery  = "query"

	keyInstrumentGroup      = "instrumentGroup"
	keyWithRepo             = "withRepo"
	keyWithoutCurrency      = "withoutCurrency"
	keyDateFrom             = "dateFrom"
	keyTicker               = "ticker"
	keySide                 = "side"
	keyFrom                 = "from"
	keyTo                   = "to"
	keyFromID               = "fromId"
	keyToID                 = "toId"
	keyQtyFrom              = "qtyFrom"
	keyQtyTo                = "qtyTo"
	keyPriceFrom            = "priceFrom"
	keyPriceTo              = "priceTo"
	keyLimit                = "limit"
	keyOffset               = "offset"
	keyTake                 = "take"
	keyDepth                = "depth"
	keyMarket               = "market"
	keyExchange             = "exchange"
	keyPortfolio            = "portfolio"
	keyStop                 = "stop"
	keySearch               = "search"
	keyCountBack            = "countBack"
	keyRiskCategoryID       = "riskCategoryId"
	keyOrderByTradeDate     = "orderByTradeDate"
	keyDescending           = "descending"
	keyUntraded             = "untraded"
	keyIncludeOld           = "includeOld"
	keyIncludeNonBaseBoards = "includeNonBaseBoards"
	keyIncludeVirtualTrades = "includeVirtualTrades"

	// History (candle) required query keys.
	keySymbol = "symbol"
	keyTf     = "tf"
)

// itoa64 renders an int64 path segment (order/stop ids) in base 10.
func itoa64(v int64) string { return strconv.FormatInt(v, 10) }

// heavyValues returns a restkit.Values pre-pinned to ?format=Heavy, selecting
// the _Heavy response variant the overlay collapsed each format oneOf down to.
// Every facade method with a _Heavy-collapsed response starts here.
//
// This is a generation invariant, not a caller-configurable default:
// models.gen.go holds only the _Heavy structs, so the pin keeps the wire shape
// matching the only type we decode into. Deliberately not a parameter.
func heavyValues() restkit.Values {
	v := restkit.NewValues()
	v.Set(keyFormat, formatHeavy)
	return v
}

// setTime sets key to the Alor date-time rendering of t when non-nil and
// non-zero, else leaves it absent. The generic restkit.Values setters cannot
// carry the package-local Time type, so date-times use the embedded Set behind
// this nil/zero check.
func setTime(v restkit.Values, key string, t *Time) restkit.Values {
	if t != nil && !t.IsZero() {
		v.Set(key, t.text())
	}
	return v
}
