package repository

import (
	"context"

	"github.com/yerobalg/wealthpulse-service/helper/db"
	"github.com/yerobalg/wealthpulse-service/helper/errors"

	"github.com/yerobalg/wealthpulse-service/src/entity"
)

type RoleInterface interface {
	GetByCode(ctx context.Context, code string) (entity.Role, error)
}

type role struct {
	db db.DB
}

func InitRole(db db.DB) RoleInterface {
	return &role{db: db}
}

func (r *role) GetByCode(ctx context.Context, code string) (entity.Role, error) {
	var result entity.Role

	res := r.db.Get(ctx).Where("code = ?", code).First(&result)
	if res.RowsAffected == 0 {
		return result, errors.NotFound("Role")
	} else if res.Error != nil {
		return result, res.Error
	}

	return result, nil
}
