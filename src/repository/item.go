package repository

import (
	"context"

	"github.com/yerobalg/wealthpulse-service/helper/db"
	"github.com/yerobalg/wealthpulse-service/helper/errors"

	"github.com/yerobalg/wealthpulse-service/src/entity"
)

type ItemInterface interface {
	Create(ctx context.Context, item *entity.Item) error
	GetByID(ctx context.Context, id int64) (entity.Item, error)
	Update(ctx context.Context, id int64, item entity.Item) error
	Delete(ctx context.Context, id int64, deletedBy int64) error
	GetList(ctx context.Context, req entity.PaginationRequest, searchColumns []string) ([]entity.Item, *entity.PaginationResponse, error)
}

type item struct {
	db db.DB
}

func InitItem(db db.DB) ItemInterface {
	return &item{db: db}
}

func (i *item) Create(ctx context.Context, item *entity.Item) error {
	return i.db.Get(ctx).Create(item).Error
}

func (i *item) GetByID(ctx context.Context, id int64) (entity.Item, error) {
	var result entity.Item

	res := i.db.Get(ctx).Where("id = ?", id).First(&result)
	if res.RowsAffected == 0 {
		return result, errors.NotFound("Item")
	} else if res.Error != nil {
		return result, res.Error
	}

	return result, nil
}

func (i *item) Update(ctx context.Context, id int64, item entity.Item) error {
	res := i.db.Get(ctx).Model(&entity.Item{}).Where("id = ?", id).Updates(&item)
	if res.RowsAffected == 0 {
		return errors.NotFound("Item")
	} else if res.Error != nil {
		return res.Error
	}

	return nil
}

func (i *item) Delete(ctx context.Context, id int64, deletedBy int64) error {
	res := i.db.Get(ctx).Model(&entity.Item{}).Where("id = ?", id).Update("deleted_by", deletedBy)
	if res.Error != nil {
		return res.Error
	}

	res = i.db.Get(ctx).Where("id = ?", id).Delete(&entity.Item{})
	if res.RowsAffected == 0 {
		return errors.NotFound("Item")
	} else if res.Error != nil {
		return res.Error
	}

	return nil
}

func (i *item) GetList(
	ctx context.Context,
	req entity.PaginationRequest,
	searchColumns []string,
) ([]entity.Item, *entity.PaginationResponse, error) {
	db := i.db.Get(ctx)

	var items []entity.Item
	dataQuery := db.Model(&entity.Item{}).Order(`"items"."created_at" DESC`)
	dataQuery = entity.ApplyPagination(dataQuery, req, searchColumns)
	if err := dataQuery.Find(&items).Error; err != nil {
		return nil, nil, err
	}

	var totalElements int64
	countQuery := db.Model(&entity.Item{})
	countQuery = entity.ApplySearchOnly(countQuery, req, searchColumns)
	if err := countQuery.Count(&totalElements).Error; err != nil {
		return nil, nil, err
	}

	pagination := entity.BuildPaginationResponse(req, int64(len(items)), totalElements)
	return items, pagination, nil
}
