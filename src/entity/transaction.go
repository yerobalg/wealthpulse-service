package entity

import (
	"gorm.io/gorm"
)

const (
	TransactionTypeBuy  = "buy"
	TransactionTypeSell = "sell"
)

type Transaction struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy" gorm:"default:null"`
	UpdatedBy *int64         `json:"updatedBy" gorm:"default:null"`
	DeletedBy *int64         `json:"deletedBy" gorm:"default:null"`

	AssetID           int64   `json:"assetId" gorm:"index;not null"`
	Type              string  `json:"type" gorm:"not null;type:text"`
	Quantity          string  `json:"quantity" gorm:"not null;type:text"`
	PriceUSD          *string `json:"priceUsd" gorm:"type:text;default:null"`
	PriceIDR          *string `json:"priceIdr" gorm:"type:text;default:null"`
	AnnualReturnBonds *string `json:"annualReturnBonds" gorm:"type:text;default:null"`
	TransactionDate   int64   `json:"transactionDate" gorm:"index;not null"`
	Notes             *string `json:"notes" gorm:"type:text;default:null"`

	Asset Asset `json:"asset" gorm:"foreignKey:AssetID"`
}
