package alor

import (
	"context"
	"net/http"
	"net/url"

	"github.com/acidsailor/restkit"
)

// portfolioService groups the Portfolio API operations. Obtain it via Client.Portfolio.
type portfolioService struct{ c *Client }

// PortfolioSummaryRequest selects the exchange/portfolio to summarize.
type PortfolioSummaryRequest struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
}

// Summary returns the portfolio summary for the given exchange/portfolio.
func (s *portfolioService) Summary(
	ctx context.Context,
	params PortfolioSummaryRequest,
) (*ResponseSummaryHeavy, error) {
	return do[*ResponseSummaryHeavy](ctx, s.c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/summary"),
		heavyValues(), nil)
}

// PortfolioPositionsRequest selects the exchange/portfolio whose positions to return.
type PortfolioPositionsRequest struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
}

// Positions returns all positions for the given exchange/portfolio.
func (s *portfolioService) Positions(
	ctx context.Context,
	params PortfolioPositionsRequest,
) (*ResponsePositionsHeavy, error) {
	return do[*ResponsePositionsHeavy](ctx, s.c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/positions"),
		heavyValues(), nil)
}

// PortfolioRiskRequest selects the exchange/portfolio whose risk metrics to return.
type PortfolioRiskRequest struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
}

// Risk returns risk metrics for the given exchange/portfolio.
func (s *portfolioService) Risk(
	ctx context.Context,
	params PortfolioRiskRequest,
) (*ResponseRiskHeavy, error) {
	return do[*ResponseRiskHeavy](ctx, s.c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/risk"),
		heavyValues(), nil)
}

// PortfolioPositionRequest selects the single-symbol position to return.
type PortfolioPositionRequest struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
	Symbol    string `json:"symbol"`
}

// Position returns the position in a single symbol for the given
// exchange/portfolio.
func (s *portfolioService) Position(
	ctx context.Context,
	params PortfolioPositionRequest,
) (*ResponsePositionHeavy, error) {
	path := clientPath(
		params.Exchange,
		params.Portfolio,
		"/positions/"+url.PathEscape(params.Symbol),
	)
	return do[*ResponsePositionHeavy](ctx, s.c, http.MethodGet,
		path, heavyValues(), nil)
}

// PortfolioPositionsByLoginRequest selects the trade-account login and an optional
// currency filter.
type PortfolioPositionsByLoginRequest struct {
	Login           string `json:"login"`
	WithoutCurrency *bool  `json:"withoutCurrency,omitempty"` // drop currency positions when true
}

// PositionsByLogin returns all positions for the given trade-account login,
// across its portfolios.
func (s *portfolioService) PositionsByLogin(
	ctx context.Context,
	params PortfolioPositionsByLoginRequest,
) (*ResponsePositionsHeavy, error) {
	path := "/md/v2/Clients/" + url.PathEscape(params.Login) + "/positions"
	q := heavyValues().Bool(keyWithoutCurrency, params.WithoutCurrency)
	return do[*ResponsePositionsHeavy](ctx, s.c, http.MethodGet, path, q, nil)
}

// PortfolioFortsRiskRequest selects the exchange/portfolio whose FORTS risk to return.
type PortfolioFortsRiskRequest struct {
	Exchange  string `json:"exchange"`
	Portfolio string `json:"portfolio"`
}

// FortsRisk returns FORTS (derivatives-market) risk metrics for the given
// exchange/portfolio.
func (s *portfolioService) FortsRisk(
	ctx context.Context,
	params PortfolioFortsRiskRequest,
) (*ResponseFortsRiskHeavy, error) {
	return do[*ResponseFortsRiskHeavy](ctx, s.c, http.MethodGet,
		clientPath(params.Exchange, params.Portfolio, "/fortsrisk"),
		heavyValues(), nil)
}

// PortfolioRiskRatesRequest selects the required risk category and the optional
// risk-rate filters.
type PortfolioRiskRatesRequest struct {
	RiskCategoryID int     `json:"riskCategoryId"`     // required risk category id
	Exchange       *string `json:"exchange,omitempty"` // exchange-code filter
	Ticker         *string `json:"ticker,omitempty"`   // ticker symbol filter
	Search         *string `json:"search,omitempty"`   // free-text search filter
	Limit          *int    `json:"limit,omitempty"`    // cap on returned items
	Offset         *int    `json:"offset,omitempty"`   // items to skip (pagination)
}

// RiskRates returns risk rates for the given risk category. RiskCategoryID
// is required.
func (s *portfolioService) RiskRates(
	ctx context.Context,
	params PortfolioRiskRatesRequest,
) (*ResponseRiskRates, error) {
	q := restkit.NewValues().
		Int(keyRiskCategoryID, &params.RiskCategoryID).
		Str(keyExchange, params.Exchange).
		Str(keyTicker, params.Ticker).
		Str(keySearch, params.Search).
		Int(keyLimit, params.Limit).
		Int(keyOffset, params.Offset)
	return do[*ResponseRiskRates](ctx, s.c, http.MethodGet,
		"/md/v2/risk/rates", q, nil)
}
