package entity

import (
	"gorm.io/gorm"
)

const (
	RoleCodeAdmin = "admin"
	RoleCodeUser  = "user"
)

type Role struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy" gorm:"default:null"`
	UpdatedBy *int64         `json:"updatedBy" gorm:"default:null"`
	DeletedBy *int64         `json:"deletedBy" gorm:"default:null"`

	Name string `json:"name" gorm:"not null;type:varchar(255)"`
	Code string `json:"code" gorm:"not null;unique;type:varchar(255)"`
}
