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
