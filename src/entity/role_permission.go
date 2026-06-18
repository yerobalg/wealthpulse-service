package entity

import (
	"gorm.io/gorm"
)

type RolePermission struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy" gorm:"default:null"`
	UpdatedBy *int64         `json:"updatedBy" gorm:"default:null"`
	DeletedBy *int64         `json:"deletedBy" gorm:"default:null"`

	RoleID       int64 `json:"roleId" gorm:"index;not null;constraint:OnDelete:CASCADE;references:id;foreignKey:RoleID"`
	PermissionID int64 `json:"permissionId" gorm:"index;not null;constraint:OnDelete:CASCADE;references:id;foreignKey:PermissionID"`
}
