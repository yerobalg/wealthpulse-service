package entity

import (
	"gorm.io/gorm"
)

type RevokedTokenReason string

const (
	RevokedTokenReasonLogout          RevokedTokenReason = "LOGOUT"
	RevokedTokenReasonExpired         RevokedTokenReason = "EXPIRED"
	RevokedTokenReasonPasswordChanged RevokedTokenReason = "PASSWORD_CHANGED"
	RevokedTokenReasonRevokedManually RevokedTokenReason = "REVOKED_MANUALLY"
)

type RevokedToken struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy" gorm:"default:null"`
	UpdatedBy *int64         `json:"updatedBy" gorm:"default:null"`
	DeletedBy *int64         `json:"deletedBy" gorm:"default:null"`

	UserID    int64              `json:"userId" gorm:"index;not null;constraint:OnDelete:CASCADE;references:id;foreignKey:UserID"`
	Token     string             `json:"token" gorm:"not null;unique;type:text"`
	ExpiredAt *int64             `json:"expiredAt"`
	Reason    RevokedTokenReason `json:"reason" gorm:"not null;type:varchar(255)"`
}
