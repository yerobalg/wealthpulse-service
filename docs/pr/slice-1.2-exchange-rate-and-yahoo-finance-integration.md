# Slice 1.2 — Exchange Rate & Yahoo Finance Integration

## Summary

Extends the `asset_price` repository with two more market-data sources, both as methods on
`AssetPriceInterface` (no separate provider files): the USD→IDR exchange rate from **Open Exchange
Rates**, and stock/ETF prices + symbol search from **Yahoo Finance** (covering US instruments by bare
ticker and IDX instruments via the `.JK` suffix). All three price sources — CoinGecko (crypto), Yahoo
(stocks/ETFs), and Open Exchange Rates (FX) — now live behind the single `AssetPriceInterface`.

See [docs/technical/wealthpulse-project-plan.md](../technical/wealthpulse-project-plan.md) for the
full plan; this PR is **Slice 1.2** — the exchange-rate and Yahoo Finance providers — within
**Slice 1 (Third-party asset integration)**, following [Slice 1.1](slice-1.1-coingecko-asset-price-integration.md).

## Changes

- **`GetUSDIDRRates` (Open Exchange Rates).** Calls `/latest.json` with `app_id`, `base=USD`,
  `symbols=IDR`, `prettyprint=false`, `show_alternative=false` through `helper/httpclient.GetJSON`,
  and maps `rates.IDR` + `timestamp` onto `entity.USDIDRRate{ Rate, Timestamp }` (rate as a decimal
  string — IDR per 1 USD).
- **`GetStockPrice` (Yahoo Finance).** Calls the keyless `/v8/finance/chart/{symbol}` endpoint with a
  browser `User-Agent` (the endpoint 403/429s without one), and maps
  `chart.result[0].meta` onto `entity.StockPrice{ Ticker, Currency, Price, Timestamp }` (price in the
  instrument's native currency — USD for US, IDR for `.JK`). One symbol per call, so the caller loops
  per asset; a single bad ticker can't fail a whole batch.
- **`SearchStock` (Yahoo Finance).** Calls the keyless `/v1/finance/search?q={ticker}` endpoint and
  filters the `quotes` array down to the requested `entity.SearchStockParam.Type` — one of `us_stock`,
  `us_etf`, `idx_stock`. A `stockSearchCriteria` map routes each type to a `{ quoteType, isIDX }` pair
  (US stock = `EQUITY` + non-`.JK`, US ETF = `ETF` + non-`.JK`, IDX stock = `EQUITY` + `.JK`); an
  unknown type returns an empty result with no call. Each match maps onto
  `entity.StockSearchResult{ Ticker, Name, Exchange, QuoteType }` (name prefers `shortname`, falls back
  to `longname`) — the metadata needed to create an asset.
- **Single repository.** All four methods are on `AssetPriceInterface` /
  `repository/asset_price.go` alongside `GetCryptoPrices`; the entity DTOs (`USDIDRRate`, `StockPrice`,
  `SearchStockParam`, `StockSearchResult`) live in `entity/asset_price.go`. The earlier standalone
  `exchange_rate` entity/repository files were folded in and removed.
- **Wiring.** `InitAssetPrice` now takes the shared `httpclient` plus three provider configs
  (`CoinGeckoConfig`, `YahooFinanceConfig`, `ExchangeRateConfig`). `main.go` reads
  `YAHOO_FINANCE_BASE_URL`, `EXCHANGERATE_BASE_URL`, and `EXCHANGERATE_API_KEY` (the OXR app id).
- **Config.** Added `EXCHANGERATE_BASE_URL` to `.env.example` (set to `https://openexchangerates.org/api`);
  `YAHOO_FINANCE_BASE_URL` was already present (set to `https://query1.finance.yahoo.com`).

## Test Plan

- [ ] `go build ./...` and `go vet ./...` are clean.
- [ ] **USD→IDR rate:** `GetUSDIDRRates` returns a `USDIDRRate` with a non-empty decimal `Rate`
      (e.g. `"17860.6"`) and a non-zero `Timestamp`; the float `17860.6` round-trips exactly as a string.
- [ ] **US stock:** `GetStockPrice(ctx, "AAPL")` returns `Currency = "USD"` and a non-empty `Price`.
- [ ] **US ETF:** `GetStockPrice(ctx, "VOO")` returns `Currency = "USD"` and a non-empty `Price`
      (ETFs and stocks share the same symbol space — no ticker-format difference).
- [ ] **IDX stock:** `GetStockPrice(ctx, "BBCA.JK")` returns `Currency = "IDR"` and a non-empty `Price`.
- [ ] **Search US stock:** `SearchStock({ Ticker: "AAPL", Type: "us_stock" })` returns Apple with
      `QuoteType = "EQUITY"` and excludes ETFs and `.JK` symbols.
- [ ] **Search US ETF:** `SearchStock({ Ticker: "VOO", Type: "us_etf" })` returns the ETF
      (`QuoteType = "ETF"`) and excludes plain equities.
- [ ] **Search IDX stock:** `SearchStock({ Ticker: "BBCA", Type: "idx_stock" })` returns only `.JK`
      equities (e.g. `BBCA.JK`), excluding US matches.
- [ ] **Search unknown type:** `SearchStock` with an unrecognized `Type` returns an empty slice and
      makes no HTTP call.
- [ ] **Missing User-Agent:** confirm Yahoo rejects requests without the `User-Agent` header (manual
      sanity check) — the repository always sends it.
- [ ] **Unknown ticker:** a symbol Yahoo doesn't recognize surfaces as an error (via
      `httpclient.StatusError`) rather than a panic, and does not abort other tickers (caller loops).
- [ ] **Malformed/empty body:** `chart.result` missing/empty yields a zero-value `StockPrice` (nil-safe
      `yahooChartMeta`) rather than a panic.
- [ ] **Provider error:** a non-2xx from either endpoint is returned as an error, never panicked.

## Checklist

- [x] One commit per layer; each builds standalone.
- [x] All outbound HTTP goes through `helper/httpclient`; no direct `net/http` in the repository.
- [x] Provider errors are returned, never panicked; nil-safe JSON extraction.
- [x] Prices/rates rendered as decimal strings, not floats.
- [x] No new API key stored in source; provider config injected from env.
