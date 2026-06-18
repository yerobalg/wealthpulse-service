package repository

import (
	"context"

	"github.com/yerobalg/wealthpulse-service/helper/db"

	"github.com/yerobalg/wealthpulse-service/src/entity"
)

type PermissionInterface interface {
	GetByRoleID(ctx context.Context, roleID int64) ([]entity.Permission, error)
}

type permission struct {
	db db.DB
}

func InitPermission(db db.DB) PermissionInterface {
	return &permission{db: db}
}

func (p *permission) GetByRoleID(ctx context.Context, roleID int64) ([]entity.Permission, error) {
	var permissions []entity.Permission

	res := p.db.Get(ctx).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id AND role_permissions.deleted_at IS NULL").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions)
	if res.Error != nil {
		return permissions, res.Error
	}

	return permissions, nil
}
