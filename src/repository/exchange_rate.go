package repository

import (
	"context"

	"github.com/yerobalg/wealthpulse-service/helper/httpclient"

	"github.com/yerobalg/wealthpulse-service/src/entity"
)

const (
	openExchangeLatestPath = "/latest.json"
	exchangeRateBaseUSD    = "USD"
	exchangeRateSymbolIDR  = "IDR"
)

// ExchangeRateConfig holds the Open Exchange Rates base URL and app id (the
// EXCHANGERATE_API_KEY env value), read from the environment and injected at
// startup.
type ExchangeRateConfig struct {
	BaseURL string
	AppID   string
}

type ExchangeRateInterface interface {
	GetUSDIDRRates(ctx context.Context) (entity.USDIDRRate, error)
}

type exchangeRate struct {
	httpClient httpclient.Interface
	config     ExchangeRateConfig
}

func InitExchangeRate(httpClient httpclient.Interface, config ExchangeRateConfig) ExchangeRateInterface {
	return &exchangeRate{httpClient: httpClient, config: config}
}

// GetUSDIDRRates fetches the latest USD→IDR rate from Open Exchange Rates'
// /latest.json and maps the decoded body onto entity.USDIDRRate.
func (e *exchangeRate) GetUSDIDRRates(ctx context.Context) (entity.USDIDRRate, error) {
	res, err := e.httpClient.GetJSON(ctx, httpclient.Request{
		URL: e.config.BaseURL + openExchangeLatestPath,
		Query: map[string]string{
			"app_id":           e.config.AppID,
			"base":             exchangeRateBaseUSD,
			"symbols":          exchangeRateSymbolIDR,
			"prettyprint":      "false",
			"show_alternative": "false",
		},
	})
	if err != nil {
		return entity.USDIDRRate{}, err
	}

	return toUSDIDRRate(res), nil
}

func toUSDIDRRate(res map[string]any) entity.USDIDRRate {
	rates, _ := res["rates"].(map[string]any)
	return entity.USDIDRRate{
		Rate:      numberStringFromJSON(rates[exchangeRateSymbolIDR]),
		Timestamp: int64FromJSON(res["timestamp"]),
	}
}

// int64FromJSON reads a JSON number (decoded as float64) as an int64.
func int64FromJSON(v any) int64 {
	f, _ := v.(float64)
	return int64(f)
}
