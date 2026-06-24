package entity

import (
	"gorm.io/gorm"
)

const (
	AssetTypeCodeCrypto        = "crypto"
	AssetTypeCodeIDX           = "idx"
	AssetTypeCodeUS            = "us"
	AssetTypeCodePreciousMetal = "precious_metal"
	AssetTypeCodeBonds         = "bonds"
	AssetTypeCodeCash          = "cash"
)

type AssetType struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy" gorm:"default:null"`
	UpdatedBy *int64         `json:"updatedBy" gorm:"default:null"`
	DeletedBy *int64         `json:"deletedBy" gorm:"default:null"`

	Name             string `json:"name" gorm:"not null;type:text"`
	Code             string `json:"code" gorm:"not null;unique;type:text"`
	TargetAllocation string `json:"targetAllocation" gorm:"not null;default:'0';type:text"`
}
