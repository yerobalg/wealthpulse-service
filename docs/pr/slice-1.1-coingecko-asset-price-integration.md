# Slice 1.1 — CoinGecko Asset Price Integration

## Summary

Adds the first market-data provider integration: an `asset_price` repository that fetches crypto
prices and searches crypto assets from CoinGecko's `/coins/markets` endpoint through the shared
`helper/httpclient` gateway. Also lands the portfolio entity structs, extends the `assets` schema
with `unique_id` (unique) and `image_url`, and teaches the HTTP gateway to decode JSON **array**
responses (CoinGecko returns an array, not an object).

See [docs/technical/wealthpulse-project-plan.md](../technical/wealthpulse-project-plan.md) for the
full plan; this PR is **Slice 1.1** — the CoinGecko provider — within **Slice 1 (Third-party asset
integration)**.

## Changes

- **`assets` schema.** Added `unique_id TEXT NOT NULL` (with a new `UNIQUE` index) and
  `image_url TEXT NULL`. The `ticker` index is now non-unique (a ticker is no longer the unique key;
  `unique_id` is).
- **Portfolio entity structs.** One file per table — `AssetType`, `Asset`, `AssetPrice`,
  `Transaction`, `Alert`, `AssetValue` — each with explicit FK fields and pointer fields for
  zero-meaningful / nullable columns, plus the code/type constants (asset-type codes, buy/sell,
  alert types).
- **`helper/httpclient` array decoding.** Added `GetJSONArray(ctx, req) ([]map[string]any, error)`.
  The internal `do` now returns the raw body, shared by `GetJSON`, `GetJSONArray`, and `PostJSON`,
  so array endpoints like CoinGecko's `/coins/markets` are supported without per-caller `net/http`.
- **`asset_price` repository — `GetCryptoPrices`.** Calls CoinGecko `/coins/markets`
  (`vs_currency=usd`). Takes `GetCryptoPricesParam{UniqueIDs, Tickers}`: `UniqueIDs` → the `ids`
  filter (fetch prices for known assets) and takes precedence; `Tickers` → the `symbols` filter
  (search by ticker) when no `UniqueIDs` are given; neither → no call. The demo API key is sent as the
  `x-cg-demo-api-key` header when configured. Decoded rows are mapped onto `entity.CryptoPrice`
  (`UniqueID`, `Ticker`, `Name`, `ImageURL`, `PriceUSD` as a decimal string).
- **Wiring.** `repository.Init` now takes `InitParam{DB, HTTPClient, CoinGecko}`; `main.go`
  constructs the shared `httpclient` (10s timeout) and reads `COINGECKO_BASE_URL` / `COINGECKO_API_KEY`.

## Test Plan

- [ ] `go build ./...` and `go vet ./...` are clean.
- [ ] `make migrate` applies the updated portfolio migration against a fresh SQLite file; `assets`
      has `unique_id` and `image_url`, a `UNIQUE` index on `unique_id`, and a non-unique index on `ticker`.
- [ ] `goose ... down` reverses the migration cleanly.
- [ ] **Prices by unique IDs:** `GetCryptoPrices` with `UniqueIDs: ["bitcoin","ethereum"]` returns one
      `CryptoPrice` per asset, each with a non-empty `PriceUSD`, `Name`, `Ticker`, and `ImageURL`.
- [ ] **Search by tickers:** `GetCryptoPrices` with `Tickers: ["btc","eth"]` (no `UniqueIDs`) returns
      matching assets via the `symbols` filter.
- [ ] **Precedence:** when both `UniqueIDs` and `Tickers` are supplied, the request uses `ids` and
      ignores `symbols`.
- [ ] **Empty input:** with neither `UniqueIDs` nor `Tickers`, no HTTP call is made and an empty slice
      is returned.
- [ ] **Auth header:** with `COINGECKO_API_KEY` set, the `x-cg-demo-api-key` header is sent; with it
      empty, no key header is sent.
- [ ] **Provider error:** a non-2xx response surfaces as an error (via `httpclient.StatusError`) and is
      returned, never panicked.
- [ ] **Price fidelity:** `current_price` is rendered as a decimal string (no float column), e.g.
      `"64231.5"`, preserving the value for the `TEXT` price columns.

## Checklist

- [x] One commit per layer; each builds standalone.
- [x] Migration uses goose `Up`/`Down` directives (SQLite dialect).
- [x] Entity FK fields declared explicitly; zero-meaningful/nullable columns use pointers.
- [x] All outbound HTTP goes through `helper/httpclient`; no direct `net/http` in the repository.
- [x] Provider errors are returned, never panicked.
- [x] Price stored/rendered as a decimal string, not a float.
