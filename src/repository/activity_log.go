package repository

import (
	"context"

	"github.com/yerobalg/wealthpulse-service/helper/db"

	"github.com/yerobalg/wealthpulse-service/src/entity"
)

type ActivityLogInterface interface {
	Create(ctx context.Context, activityLog entity.ActivityLog) error
}

type activityLog struct {
	db db.DB
}

func InitActivityLog(db db.DB) ActivityLogInterface {
	return &activityLog{db: db}
}

func (a *activityLog) Create(ctx context.Context, activityLog entity.ActivityLog) error {
	res := a.db.Get(ctx).Create(&activityLog)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
