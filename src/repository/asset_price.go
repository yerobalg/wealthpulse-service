package repository

import (
	"context"
	"strconv"
	"strings"

	"github.com/yerobalg/wealthpulse-service/helper/httpclient"

	"github.com/yerobalg/wealthpulse-service/src/entity"
)

const (
	coinGeckoMarketsPath  = "/coins/markets"
	coinGeckoVsCurrency   = "usd"
	coinGeckoAPIKeyHeader = "x-cg-demo-api-key"

	yahooChartPath = "/v8/finance/chart/"
	// yahooUserAgent — Yahoo's unofficial chart endpoint rejects requests without
	// a browser-like User-Agent (403/429), so every call sends one.
	yahooUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

// CoinGeckoConfig holds the crypto provider base URL and demo API key, read from
// the environment and injected at startup.
type CoinGeckoConfig struct {
	BaseURL string
	APIKey  string
}

// YahooFinanceConfig holds the Yahoo Finance base URL (e.g.
// https://query1.finance.yahoo.com), used for US stocks/ETFs and IDX stocks.
// Yahoo's chart endpoint needs no API key.
type YahooFinanceConfig struct {
	BaseURL string
}

type AssetPriceInterface interface {
	GetCryptoPrices(ctx context.Context, param entity.GetCryptoPricesParam) ([]entity.CryptoPrice, error)
	GetStockPrice(ctx context.Context, ticker string) (entity.StockPrice, error)
}

type assetPrice struct {
	httpClient httpclient.Interface
	coinGecko  CoinGeckoConfig
	yahoo      YahooFinanceConfig
}

func InitAssetPrice(
	httpClient httpclient.Interface,
	coinGecko CoinGeckoConfig,
	yahoo YahooFinanceConfig,
) AssetPriceInterface {
	return &assetPrice{httpClient: httpClient, coinGecko: coinGecko, yahoo: yahoo}
}

// GetCryptoPrices fetches crypto market data from CoinGecko's /coins/markets.
// UniqueIDs are sent as the "ids" filter (fetch prices for known assets) and
// take precedence; when only Tickers are supplied they are sent as the
// "symbols" filter (search by ticker). With neither, no call is made.
func (a *assetPrice) GetCryptoPrices(
	ctx context.Context,
	param entity.GetCryptoPricesParam,
) ([]entity.CryptoPrice, error) {
	query := map[string]string{"vs_currency": coinGeckoVsCurrency}

	switch {
	case len(param.UniqueIDs) > 0:
		query["ids"] = strings.Join(param.UniqueIDs, ",")
	case len(param.Tickers) > 0:
		query["symbols"] = strings.Join(param.Tickers, ",")
	default:
		return []entity.CryptoPrice{}, nil
	}

	headers := map[string]string{}
	if a.coinGecko.APIKey != "" {
		headers[coinGeckoAPIKeyHeader] = a.coinGecko.APIKey
	}

	items, err := a.httpClient.GetJSONArray(ctx, httpclient.Request{
		URL:     a.coinGecko.BaseURL + coinGeckoMarketsPath,
		Headers: headers,
		Query:   query,
	})
	if err != nil {
		return nil, err
	}

	prices := make([]entity.CryptoPrice, 0, len(items))
	for _, item := range items {
		prices = append(prices, toCryptoPrice(item))
	}
	return prices, nil
}

// GetStockPrice fetches one instrument's latest price from Yahoo Finance's
// /v8/finance/chart/{symbol} and maps the decoded body onto entity.StockPrice.
// US symbols use their bare ticker (AAPL); IDX symbols use the ".JK" suffix
// (BBCA.JK), and the native currency comes back in the response.
//
// The endpoint serves one symbol per call, so the caller (scheduler) loops over
// its assets — keeping a single failed ticker from failing the whole batch.
func (a *assetPrice) GetStockPrice(ctx context.Context, ticker string) (entity.StockPrice, error) {
	res, err := a.httpClient.GetJSON(ctx, httpclient.Request{
		URL:     a.yahoo.BaseURL + yahooChartPath + ticker,
		Headers: map[string]string{"User-Agent": yahooUserAgent},
	})
	if err != nil {
		return entity.StockPrice{}, err
	}

	return toStockPrice(ticker, res), nil
}

func toCryptoPrice(item map[string]any) entity.CryptoPrice {
	return entity.CryptoPrice{
		UniqueID: stringFromJSON(item["id"]),
		Ticker:   strings.ToUpper(stringFromJSON(item["symbol"])),
		Name:     stringFromJSON(item["name"]),
		ImageURL: stringFromJSON(item["image"]),
		PriceUSD: numberStringFromJSON(item["current_price"]),
	}
}

func toStockPrice(ticker string, res map[string]any) entity.StockPrice {
	meta := yahooChartMeta(res)
	return entity.StockPrice{
		Ticker:    ticker,
		Currency:  stringFromJSON(meta["currency"]),
		Price:     numberStringFromJSON(meta["regularMarketPrice"]),
		Timestamp: int64FromJSON(meta["regularMarketTime"]),
	}
}

// yahooChartMeta digs out chart.result[0].meta, returning an empty map when any
// step is missing so the caller's lookups stay nil-safe.
func yahooChartMeta(res map[string]any) map[string]any {
	chart, _ := res["chart"].(map[string]any)
	results, _ := chart["result"].([]any)
	if len(results) == 0 {
		return map[string]any{}
	}

	first, _ := results[0].(map[string]any)
	meta, _ := first["meta"].(map[string]any)
	return meta
}

func stringFromJSON(v any) string {
	s, _ := v.(string)
	return s
}

// numberStringFromJSON renders a JSON number (decoded as float64) as a decimal
// string so it can be stored in the TEXT price columns without losing the value
// to a float column.
func numberStringFromJSON(v any) string {
	switch n := v.(type) {
	case float64:
		return strconv.FormatFloat(n, 'f', -1, 64)
	case string:
		return n
	default:
		return ""
	}
}
