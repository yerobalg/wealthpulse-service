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
)

// CoinGeckoConfig holds the provider base URL and demo API key, read from the
// environment and injected at startup.
type CoinGeckoConfig struct {
	BaseURL string
	APIKey  string
}

type AssetPriceInterface interface {
	GetCryptoPrices(ctx context.Context, param entity.GetCryptoPricesParam) ([]entity.CryptoPrice, error)
}

type assetPrice struct {
	httpClient httpclient.Interface
	config     CoinGeckoConfig
}

func InitAssetPrice(httpClient httpclient.Interface, config CoinGeckoConfig) AssetPriceInterface {
	return &assetPrice{httpClient: httpClient, config: config}
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
	if a.config.APIKey != "" {
		headers[coinGeckoAPIKeyHeader] = a.config.APIKey
	}

	items, err := a.httpClient.GetJSONArray(ctx, httpclient.Request{
		URL:     a.config.BaseURL + coinGeckoMarketsPath,
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

func toCryptoPrice(item map[string]any) entity.CryptoPrice {
	return entity.CryptoPrice{
		UniqueID: stringFromJSON(item["id"]),
		Ticker:   strings.ToUpper(stringFromJSON(item["symbol"])),
		Name:     stringFromJSON(item["name"]),
		ImageURL: stringFromJSON(item["image"]),
		PriceUSD: numberStringFromJSON(item["current_price"]),
	}
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
