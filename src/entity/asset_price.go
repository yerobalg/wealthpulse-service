package entity

import (
	"gorm.io/gorm"
)

type AssetPrice struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	AssetID  int64   `json:"assetId" gorm:"index;not null"`
	PriceUSD *string `json:"priceUsd" gorm:"type:text;default:null"`
	PriceIDR *string `json:"priceIdr" gorm:"type:text;default:null"`

	Asset Asset `json:"asset" gorm:"foreignKey:AssetID"`
}

// GetCryptoPricesParam selects which crypto assets to fetch from the provider.
// UniqueIDs maps to the provider "ids" filter (fetch prices for known assets);
// Tickers maps to the provider "symbols" filter (search assets by ticker). When
// both are supplied UniqueIDs wins and Tickers is ignored.
type GetCryptoPricesParam struct {
	UniqueIDs []string
	Tickers   []string
}

// CryptoPrice is one crypto asset's provider market data, carrying both the
// metadata needed to create an asset and its current USD price.
type CryptoPrice struct {
	UniqueID string `json:"uniqueId"`
	Ticker   string `json:"ticker"`
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl"`
	PriceUSD string `json:"priceUsd"`
}

// StockPrice is one instrument's latest price from Yahoo Finance, covering both
// US stocks/ETFs (currency USD) and IDX stocks (ticker suffix ".JK", currency
// IDR). Price is in the instrument's native currency as a decimal string;
// Timestamp is the provider's quote time in epoch seconds.
type StockPrice struct {
	Ticker    string `json:"ticker"`
	Currency  string `json:"currency"`
	Price     string `json:"price"`
	Timestamp int64  `json:"timestamp"`
}

// USDIDRRate is the latest USD→IDR rate from the exchange-rate provider
// (Open Exchange Rates). Rate is IDR per 1 USD as a decimal string; Timestamp
// is the provider's quote time in epoch seconds.
type USDIDRRate struct {
	Rate      string `json:"rate"`
	Timestamp int64  `json:"timestamp"`
}

// Stock search types — which instrument class to keep from the provider's
// symbol search (a ticker matches both EQUITY and ETF results across markets).
const (
	StockSearchTypeUSStock  = "us_stock"
	StockSearchTypeUSETF    = "us_etf"
	StockSearchTypeIDXStock = "idx_stock"
)

// SearchStockParam searches Yahoo Finance for instruments matching Ticker,
// keeping only those of the requested Type (one of the StockSearchType* values).
type SearchStockParam struct {
	Ticker string
	Type   string
}

// StockSearchResult is one instrument returned by the provider symbol search,
// carrying the metadata needed to create an asset. QuoteType is the provider's
// own classification (e.g. "EQUITY", "ETF").
type StockSearchResult struct {
	Ticker    string `json:"ticker"`
	Name      string `json:"name"`
	Exchange  string `json:"exchange"`
	QuoteType string `json:"quoteType"`
}
