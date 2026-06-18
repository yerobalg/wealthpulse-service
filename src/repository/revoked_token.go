package repository

import (
	"context"

	"github.com/yerobalg/wealthpulse-service/helper/db"

	"github.com/yerobalg/wealthpulse-service/src/entity"
)

type RevokedTokenInterface interface {
	Create(ctx context.Context, revokedToken entity.RevokedToken) error
	IsTokenRevoked(ctx context.Context, token string) (bool, error)
}

type revokedToken struct {
	db db.DB
}

func InitRevokedToken(db db.DB) RevokedTokenInterface {
	return &revokedToken{db: db}
}

func (r *revokedToken) Create(ctx context.Context, revokedToken entity.RevokedToken) error {
	res := r.db.Get(ctx).Create(&revokedToken)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *revokedToken) IsTokenRevoked(ctx context.Context, token string) (bool, error) {
	var exists bool

	err := r.db.Get(ctx).Model(&entity.RevokedToken{}).
		Select("1").
		Where("token = ?", token).
		Limit(1).
		Find(&exists).Error
	if err != nil {
		return false, err
	}

	return exists, nil
}
