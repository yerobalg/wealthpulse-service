package entity

import (
	"gorm.io/gorm"
)

const (
	AlertTypeUpper = "upper"
	AlertTypeLower = "lower"
	AlertTypePct   = "pct"
)

type Alert struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy" gorm:"default:null"`
	UpdatedBy *int64         `json:"updatedBy" gorm:"default:null"`
	DeletedBy *int64         `json:"deletedBy" gorm:"default:null"`

	AssetID         *int64 `json:"assetId" gorm:"index;default:null"`
	Type            string `json:"type" gorm:"not null;type:text"`
	Threshold       string `json:"threshold" gorm:"not null;type:text"`
	IsActive        *bool  `json:"isActive" gorm:"not null;default:true"`
	LastTriggeredAt *int64 `json:"lastTriggeredAt" gorm:"default:null"`

	Asset Asset `json:"asset" gorm:"foreignKey:AssetID"`
}
