package repository

import (
	"context"

	"github.com/yerobalg/wealthpulse-service/helper/db"
	"github.com/yerobalg/wealthpulse-service/helper/errors"

	"github.com/yerobalg/wealthpulse-service/src/entity"
)

type UserInterface interface {
	Get(ctx context.Context, req entity.UserRequest) (entity.User, error)
	Update(ctx context.Context, req entity.UserRequest, user entity.User) error
	Create(ctx context.Context, user *entity.User) error
	GetListWithRole(ctx context.Context, req entity.PaginationRequest, searchColumns []string) ([]entity.User, *entity.PaginationResponse, error)
}

type user struct {
	db db.DB
}

func InitUser(db db.DB) UserInterface {
	return &user{db: db}
}

func (u *user) Get(ctx context.Context, req entity.UserRequest) (entity.User, error) {
	var user entity.User

	whereClause := u.db.GetWhereClauseFromParamTag(req)

	res := u.db.Get(ctx).InnerJoins("Role").Where(whereClause).First(&user)
	if res.RowsAffected == 0 {
		return user, errors.NotFound("User")
	} else if res.Error != nil {
		return user, res.Error
	}

	return user, nil
}

func (u *user) Create(ctx context.Context, user *entity.User) error {
	return u.db.Get(ctx).Create(user).Error
}

func (u *user) Update(ctx context.Context, req entity.UserRequest, user entity.User) error {
	whereClause := u.db.GetWhereClauseFromParamTag(req)

	res := u.db.Get(ctx).Where(whereClause).Updates(&user)
	if res.RowsAffected == 0 {
		return errors.NotFound("User")
	} else if res.Error != nil {
		return res.Error
	}

	return nil
}

func (u *user) GetListWithRole(
	ctx context.Context,
	req entity.PaginationRequest,
	searchColumns []string,
) ([]entity.User, *entity.PaginationResponse, error) {
	db := u.db.Get(ctx)

	var users []entity.User
	dataQuery := db.Model(&entity.User{}).
		Joins("Role").
		Order(`"Role"."name" ASC`).
		Order(`"users"."name" ASC`)
	dataQuery = entity.ApplyPagination(dataQuery, req, searchColumns)
	if err := dataQuery.Find(&users).Error; err != nil {
		return nil, nil, err
	}

	var totalElements int64
	countQuery := db.Model(&entity.User{})
	countQuery = entity.ApplySearchOnly(countQuery, req, searchColumns)
	if err := countQuery.Count(&totalElements).Error; err != nil {
		return nil, nil, err
	}

	pagination := entity.BuildPaginationResponse(req, int64(len(users)), totalElements)
	return users, pagination, nil
}
