package entity

import (
	"gorm.io/gorm"
)

type AssetValue struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	AssetTypeID   int64   `json:"assetTypeId" gorm:"index;not null"`
	TotalValueUSD *string `json:"totalValueUsd" gorm:"type:text;default:null"`
	TotalValueIDR *string `json:"totalValueIdr" gorm:"type:text;default:null"`

	AssetType AssetType `json:"assetType" gorm:"foreignKey:AssetTypeID"`
}
