# CLAUDE.md

Go HTTP client for the [Alor broker API](https://alor.dev/docs/). Module
`github.com/acidsailor/alor` (Go 1.26). A curated facade over the shared
[`restkit`](https://github.com/acidsailor/restkit) HTTP-JSON transport; response
models are code-generated from Alor's OpenAPI spec by
[oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) (types only — plain
structs, no generated client/server). Scaffolded from
`gh:acidsailor/go-scaffolds` (project_type: library); `.copier-answers.yml` is
Copier-managed — never edit it by hand.

## Common commands

Build/test/lint run through [Task](https://taskfile.dev) (`taskfile.yml`):

```sh
task test     # go test ./...
task lint     # golangci-lint fmt + golangci-lint run --fix (mutates files)
task ci       # read-only verification: golangci-lint fmt --diff + run (no autofix)
task check    # composite: lint (mutating) + test
task tidy     # go mod tidy
```

Plain Go also works: `go build ./...`, `go test ./...`. Lint requires
`golangci-lint` on PATH (config in `.golangci.yaml`: standard linters +
`modernize`; formatters `gofumpt` with extra-rules and `golines` at max-len 80).

CI (`.github/workflows/ci.yml`) delegates to the reusable
`acidsailor/go-scaffolds/.github/workflows/go-ci.yml@v1` workflow.

No `main` package — this is a library; "run" means importing it from a consumer.
The entry point is `auth.New(...)` → `alor.NewClient(...)`.

### Code generation

The spec/codegen tasks shell out to Docker (Speakeasy, Redocly) and `go run`:

```sh
task spec      # curl upstream spec -> spec/spec-upstream.yml (committed verbatim)
task overlay   # speakeasy overlay apply --strict: spec-upstream + overlay -> spec/spec.yml
task gen       # go run oapi-codegen@v2.5.0 -> models.gen.go (deps overlay; fails on any warning)
task deref     # redocly bundle --dereferenced -> spec/spec-deref.json (deps overlay; MCP branch)
```

Do not edit `models.gen.go` by hand — `task gen` regenerates it wholesale
(config in `oapi-codegen.yaml`). After a spec change rerun both `gen` and `deref`
(independent branches off `spec/spec.yml`), then `task tidy` if imports shifted.

## Architecture

Single public package `alor` (top-level files) plus two subpackages: `auth/` and
`spec/`. No `main`.

### Transport + facade (`alor` package)

- `client.go` — `Client` wraps one `*restkit.Client`; immutable after
  construction, concurrency-safe. Entry point `NewClient(endpoint, ts
  oauth2.TokenSource, opts ...Option)`. Use `EndpointProduction`
  (`https://api.alor.ru`) / `EndpointTest` (`https://apidev.alor.ru`). Options:
  `WithHTTPClient` (nil → restkit default: 30s timeout, stdlib transport,
  otelhttp-wrapped) and `WithInitialToken` (seed the `oauth2.ReuseTokenSource`
  cache). `NewClient` wraps `ts` in `oauth2.ReuseTokenSource` and installs the
  bearer-auth hook. Two generic transport helpers sit over `restkit.Do`: `do[T]`
  decodes the 2xx JSON body into `T`; `exec` backs endpoints replying with a bare
  text/plain `"success"` (issues `do[json.RawMessage]` and treats the expected
  `OpUnmarshal` failure on the non-JSON body as success).
- Alor specifics ride on restkit request hooks: `bearerAuth(ts)` (client-wide)
  fetches the token per request and sets `Authorization: Bearer …`;
  `withReqID(id)` (per-call) sets the `X-REQID` idempotency header for order
  commands (empty id → no-op).
- `wrappers.go` — the **read** facade (GET): `ServerTime`, `GetOrders`,
  `SearchSecurities`, `GetTradeHistory`, `GetOrderBook`, `GetHistory`, etc.
- `trading.go` — the **mutating** facade (POST/PUT/DELETE): order
  place/replace/cancel, `EstimateOrder(s)`, order-group CRUD.
- `params.go` — query-key constants, `heavyValues()` (pins `?format=Heavy`),
  `setTime` helper, `itoa64` path helper. Query building uses restkit's fluent
  `Values` setters (`.Str`/`.Bool`/`.Int`/`.Int64`).
- `scalars.go` — hand-written `Time` type (lenient `UnmarshalJSON` for Alor's
  zone-less timestamps + the `9999-…` .NET MaxValue "never expires" sentinel;
  `NeverExpires()` detects it).
- `errors.go` — re-exports restkit's three error types as aliases
  (`ResponseError`, `RequestError`, `ConfigError`) and the `Op*` constants.

### Facade method convention

Each method takes `ctx` plus a single `XxxParams` struct: required fields are
values, optional query filters are pointers (`*string`/`*int`/`*bool`/`*Time`,
nil → omitted). Return types are the generated structs **directly** under their
oapi-codegen names (e.g. `GetOrders` returns `ResponseOrdersHeavy`); no
type-alias layer. Mutating methods carry a `ReqID string` field (X-REQID
idempotency key) in their params; bare-`"success"` endpoints return only `error`.
Only non-deprecated operations are exposed.

### Error model

No sentinel errors — the typed error IS the category, matched with `errors.As`:
- `*ConfigError` — `NewClient` misuse (nil `TokenSource`, empty endpoint).
- `*ResponseError` — a non-2xx response (carries `StatusCode` + raw `Body`); the
  typed result of the call, not a `RequestError`.
- `*RequestError` — any other per-call failure; `Op` names the stage
  (`OpHook`/`OpMarshal`/`OpBuild`/`OpSend`/`OpRead`/`OpUnmarshal`).

Each wraps its cause via `%w`, so `errors.As`/`errors.Is` still reach a transport
`*url.Error`, `context.Canceled`/`DeadlineExceeded`, a JSON decode error, or the
auth package's restkit error (through the `OpHook` bearer-auth failure). 4xx
error bodies are surfaced as the raw `ResponseError.Body` string — no typed
error DTOs.

### `auth/` package

`auth.New(oauthHost, refreshToken, opts ...Option)` returns a
`*auth.TokenSource` (implements `oauth2.TokenSource`) that mints JWTs via Alor's
non-RFC `/refresh` endpoint. Use `auth.HostProduction`
(`https://oauth.alor.ru`) / `auth.HostTest`. Options: `auth.WithHTTPClient`,
`auth.WithAllowedPortfolios` (scope the JWT). Built on its **own** restkit client
(name `"alor-auth"`) — separate from the API client, so the two never share a
connection pool; keep that separation. `Token()` hits `/refresh` every call;
`oauth2.ReuseTokenSource` (applied by `NewClient`) caches until the JWT `exp`.
Errors are the **same** three restkit types, re-exported as `auth` aliases
(`auth.ConfigError`/`ResponseError`/`RequestError`); no sentinels.

### `spec/` package

A leaf package (`package spec`, imports only `embed`) that embeds the
dereferenced spec as `SpecDerefJSON []byte`, consumed by a downstream Alor MCP
server. `spec/spec_test.go` guards its invariants (non-empty, valid JSON,
populated paths/components, no residual `$ref`). The ref-based `spec.yml` is the
oapi-codegen input only, not embedded.

## Conventions & gotchas

- `models.gen.go` is generated — never hand-edit; rerun `task gen`.
- `task gen` fails the build on any oapi-codegen `warn` diagnostic.
- The generated file imports nothing beyond stdlib + `udecimal`/`google/uuid`
  (typed scalars via overlay `x-go-type`) and **no oapi-codegen runtime** — a
  test guards this. The overlay collapses each response-format `oneOf` to its
  `_Heavy` superset (selected by `?format=Heavy`, pinned by the method) and drops
  the WebSocket `trace:` ops + deprecated operations so their schemas are pruned.
- OTel: the client transport is otelhttp-wrapped via the **global** providers
  (`otel.GetTracerProvider`/`GetMeterProvider`); there is no per-client option —
  wire the globals at process start.
- `auth.WithHTTPClient`'s client MUST NOT be wrapped with an `oauth2.Transport`
  backed by the same source (infinite refresh recursion).
- Out of scope: WebSocket/streaming, `?format=Slim`/`Simple`, OAuth code-flow
  bootstrap (callers use `oauth2.Config.Exchange` and feed the `refresh_token`
  to `auth.New`), GraphQL.
