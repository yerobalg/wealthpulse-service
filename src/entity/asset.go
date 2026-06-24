package entity

import (
	"gorm.io/gorm"
)

type Asset struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy" gorm:"default:null"`
	UpdatedBy *int64         `json:"updatedBy" gorm:"default:null"`
	DeletedBy *int64         `json:"deletedBy" gorm:"default:null"`

	AssetTypeID int64   `json:"assetTypeId" gorm:"index;not null"`
	Name        string  `json:"name" gorm:"not null;type:text"`
	Ticker      string  `json:"ticker" gorm:"index;not null;type:text"`
	UniqueID    string  `json:"uniqueId" gorm:"unique;not null;type:text"`
	ImageURL    *string `json:"imageUrl" gorm:"type:text;default:null"`
	ExternalID  *string `json:"externalId" gorm:"type:text;default:null"`

	AssetType AssetType `json:"assetType" gorm:"foreignKey:AssetTypeID"`
}
