package entity

import (
	"gorm.io/gorm"

	"github.com/yerobalg/wealthpulse-service/helper/logger"
)

type ActivityLog struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy" gorm:"default:null"`
	UpdatedBy *int64         `json:"updatedBy" gorm:"default:null"`
	DeletedBy *int64         `json:"deletedBy" gorm:"default:null"`

	UserID           int64   `json:"userId" gorm:"index;not null;constraint:OnDelete:NO ACTION;references:id;foreignKey:UserID"`
	UserToken        string  `json:"userToken" gorm:"not null;type:text"`
	Metadata         string  `json:"metadata" gorm:"not null;type:text"`
	ActivityEvent    string  `json:"activityEvent" gorm:"not null;type:text"`
	ActivityName     string  `json:"activityName" gorm:"not null;type:text"`
	AdditionalFields *string `json:"additionalFields" gorm:"type:text;default:null"`

	User User `json:"user" gorm:"foreignKey:UserID"`
}

type ActivityLogInsertRequest struct {
	ActivityEvent    logger.AuditLogEvent
	ActivityName     string
	AdditionalFields any
}
