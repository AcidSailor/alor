# alor

Go HTTP client for the [Alor broker API](https://alor.dev/docs/).

A curated facade over the shared
[`restkit`](https://github.com/acidsailor/restkit) HTTP-JSON transport, exposing
typed methods for Alor's REST endpoints — quotes, securities, orders, positions,
trades, and order management. Response models are code-generated from Alor's
OpenAPI spec with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)
(types only — plain structs, no generated client), and the package pulls in no
code-generator runtime. Bearer-token auth (auto-refreshing) and the X-REQID order
idempotency header are layered on via restkit request hooks; the HTTP transport
is OpenTelemetry-instrumented.

**Out of scope:** WebSocket / streaming, the `?format=Slim` / `Simple` response
variants (the facade requests `Heavy` only), OAuth code-flow bootstrap, GraphQL.

## Features

- Typed methods for Alor's REST surface: market data (quotes, order book,
  candles, trades), client info (orders, positions, summary, risk), trading
  (place/replace/cancel market, limit, stop, and stop-limit orders), order
  estimation, and order-group management.
- Auto-refreshing `oauth2.TokenSource` for Alor's non-RFC `/refresh` endpoint
  (`auth` package), composable with `golang.org/x/oauth2` for third-party apps.
- Typed errors matched with `errors.As` (no sentinels), with the underlying
  cause preserved via `%w`.
- OpenTelemetry tracing + metrics on the HTTP transport via the global providers.
- Exact decimals (`udecimal`) and strict UUIDs — no lossy `float64` decode.

## Install

```sh
go get github.com/acidsailor/alor
```

Requires Go 1.26+.

## Quickstart

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/acidsailor/alor"
	"github.com/acidsailor/alor/auth"
)

func main() {
	ts, err := auth.New(auth.HostProduction, os.Getenv("ALOR_REFRESH_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	c, err := alor.NewClient(alor.EndpointProduction, ts)
	if err != nil {
		log.Fatal(err)
	}

	query := "SBER"
	securities, err := c.MarketData.Search(context.Background(),
		alor.MarketDataSearchRequest{Query: &query})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", securities)
}
```

Operations are grouped into service facades on the client (`c.Orders`, `c.StopOrders`, `c.OrderGroups`, `c.Portfolio`, `c.Trades`,
`c.MarketData`). Each method takes a context plus a single `XxxRequest` struct
— required fields as values, optional query filters as pointers (nil omits the
parameter) — and returns a pointer to the response (`*T`; collections are
`[]T`). `c.ServerTime` stays a top-level method.

### Placing an order

Order-command methods carry a `ReqID` (the `X-REQID` idempotency key): mint a
unique value per command and reuse it on retries so a resend cannot
double-submit.

```go
resp, err := c.Orders.PlaceLimit(ctx, alor.OrdersPlaceLimitRequest{
	ReqID: uuid.NewString(),
	Order: order, // alor.OrdersActionsLimitTVPost
})
```

## Authentication

`auth.New` takes the long-lived refresh token from
[alor.dev/myapps](https://alor.dev/myapps) and returns an
[`oauth2.TokenSource`](https://pkg.go.dev/golang.org/x/oauth2#TokenSource) that
mints access tokens via Alor's `/refresh` endpoint. Access tokens live ~30
minutes; `NewClient` automatically wraps the source in
[`oauth2.ReuseTokenSource`](https://pkg.go.dev/golang.org/x/oauth2#ReuseTokenSource),
caching minted tokens until the JWT `exp` elapses.

For apps acting on behalf of other Alor users, the `/authorize` + `/token` flow
is RFC 6749 compliant — bootstrap with standard `golang.org/x/oauth2` and feed
the returned `refresh_token` into `auth.New`:

```go
oauthTok, _ := oauthCfg.Exchange(ctx, code)
ts, err := auth.New(auth.HostProduction, oauthTok.RefreshToken)
```

Any `oauth2.TokenSource` works — e.g. `oauth2.StaticTokenSource` for a
short-lived access token.

## Configuration

### Hosts

| Constant | Value | |
| --- | --- | --- |
| `alor.EndpointProduction` | `https://api.alor.ru` | API |
| `alor.EndpointTest` | `https://apidev.alor.ru` | API (separate refresh token) |
| `auth.HostProduction` | `https://oauth.alor.ru` | OAuth |
| `auth.HostTest` | `https://oauthdev.alor.ru` | OAuth |

The test environment requires a separate refresh token from the
[Alor dev cabinet](https://alor.dev/).

### Options

`NewClient` and `auth.New` accept functional options. By default the HTTP client
has a 30s timeout and the stdlib transport.

- `alor.WithHTTPClient(*http.Client)` — custom round-tripper / timeout / proxy
  (nil falls back to the default).
- `alor.WithInitialToken(*oauth2.Token)` — seed the refresh cache from a
  persisted token to skip the first `/refresh` (nil is a no-op).
- `auth.WithHTTPClient(*http.Client)` — same, for the `/refresh` transport. The
  client MUST NOT be wrapped with an `oauth2.Transport` backed by the same source
  (infinite refresh recursion).
- `auth.WithAllowedPortfolios(...string)` — scope the issued JWT to specific
  portfolio IDs.

```go
c, err := alor.NewClient(alor.EndpointProduction, ts,
	alor.WithHTTPClient(&http.Client{Timeout: 60 * time.Second}),
	alor.WithInitialToken(persisted),
)
```

The resulting `*alor.Client` and `*auth.TokenSource` are immutable and safe for
concurrent use.

### Environment

The library reads no environment variables. The quickstart reads
`ALOR_REFRESH_TOKEN` only as an example source for the refresh token —
substitute your own secret loading.

## Errors

Failures are the `restkit` typed errors, re-exported as aliases so you match
them with `errors.As` (no sentinels). The cause is preserved via `%w`.

```go
var ce *alor.ConfigError   // NewClient misuse (nil TokenSource, empty endpoint)
errors.As(err, &ce)

var re *alor.ResponseError // non-2xx response
if errors.As(err, &re) { _ = re.StatusCode; _ = re.Body }

var qe *alor.RequestError  // any other per-call failure
if errors.As(err, &qe) { _ = qe.Op } // "hook", "send", "unmarshal", ...

errors.Is(err, context.DeadlineExceeded) // wrapped cause survives unwrap
```

The `auth` package re-exports the same three types as its own aliases
(`auth.ConfigError` / `auth.ResponseError` / `auth.RequestError`); a non-2xx
`/refresh` (e.g. a 401/403 for a bad or expired refresh token) is an
`*auth.ResponseError`, reachable through a bearer-hook `*alor.RequestError`.

4xx error bodies are surfaced as the raw `ResponseError.Body` string, not typed
DTOs — unmarshal it yourself if you need structured error payloads.

## OpenTelemetry

The client's HTTP transport is wrapped with
[`otelhttp`](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp),
emitting client spans and HTTP metrics through the globally configured providers
(`otel.SetTracerProvider` / `otel.SetMeterProvider`). With nothing configured the
no-op provider drops them. There is no per-client option — wire the globals once
at process start.

## Regenerating the models

The response structs live in `models.gen.go`, produced by oapi-codegen (models
only) from a spec assembled via a declarative
[OpenAPI Overlay](https://github.com/OAI/Overlay-Specification) (`spec/overlay.yml`).
The spec stages run through Docker; generation is a `go run`:

```sh
task spec      # refresh spec/spec-upstream.yml from upstream Alor
task overlay   # apply spec/overlay.yml -> spec/spec.yml (Speakeasy, --strict)
task gen       # oapi-codegen -> models.gen.go (fails on any warning)
task deref     # redocly bundle --dereferenced -> spec/spec-deref.json
task check     # lint + test
```

`models.gen.go` is generated — do not edit it by hand.

## Disclaimer

This library is provided "as is", without warranty of any kind. The author
assumes **no financial, legal, or other liability** for any losses, damages, or
consequences arising from the use of this library, including but not limited to
losses incurred through trading, order placement, or interaction with the Alor
API.

Nothing in this library, its documentation, or examples constitutes **investment
advice, a recommendation, or solicitation** to buy or sell any financial
instrument. All trading decisions are solely the responsibility of the user.
Consult a licensed financial advisor before making investment decisions.

## Отказ от ответственности

Библиотека предоставляется «как есть», без каких-либо гарантий. Автор **не несёт
финансовой, юридической или иной ответственности** за любые убытки, ущерб или
последствия, возникшие в результате использования этой библиотеки, включая, но
не ограничиваясь, убытки от торговли, выставления ордеров или взаимодействия с
API Alor.

Ничто в этой библиотеке, её документации или примерах **не является
индивидуальной инвестиционной рекомендацией**, предложением или побуждением к
покупке или продаже каких-либо финансовых инструментов. Все торговые решения
принимаются пользователем самостоятельно и под его ответственность. Перед
принятием инвестиционных решений проконсультируйтесь с лицензированным
финансовым советником.

## License

[GNU AGPL v3](./LICENSE).
