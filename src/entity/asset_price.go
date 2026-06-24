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
