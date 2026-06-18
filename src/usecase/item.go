package usecase

import (
	"context"

	"github.com/yerobalg/wealthpulse-service/helper/authcontext"
	errorLib "github.com/yerobalg/wealthpulse-service/helper/errors"
	"github.com/yerobalg/wealthpulse-service/helper/logger"
	"github.com/yerobalg/wealthpulse-service/helper/types"
	"github.com/yerobalg/wealthpulse-service/helper/validator"

	"github.com/yerobalg/wealthpulse-service/src/entity"
	"github.com/yerobalg/wealthpulse-service/src/entity/validation"
	"github.com/yerobalg/wealthpulse-service/src/repository"
)

type ItemInterface interface {
	Create(ctx context.Context, req entity.CreateItemRequest) (entity.ItemResponse, error)
	GetByID(ctx context.Context, req entity.ItemRequest) (entity.ItemResponse, error)
	Update(ctx context.Context, req entity.UpdateItemRequest) error
	Delete(ctx context.Context, req entity.ItemRequest) error
	GetList(ctx context.Context, req entity.GetListItemRequest) ([]entity.ItemResponse, *entity.PaginationResponse, error)
}

type item struct {
	itemRepo    repository.ItemInterface
	validator   validator.Interface
	activityLog ActivityLogInterface
}

type ItemInitParam struct {
	ItemRepo    repository.ItemInterface
	Validator   validator.Interface
	ActivityLog ActivityLogInterface
}

func InitItem(param ItemInitParam) ItemInterface {
	return &item{
		itemRepo:    param.ItemRepo,
		validator:   param.Validator,
		activityLog: param.ActivityLog,
	}
}

func (i *item) Create(ctx context.Context, req entity.CreateItemRequest) (entity.ItemResponse, error) {
	if err := i.validator.Bind(req, validation.ItemCreate); err != nil {
		return entity.ItemResponse{}, err
	}

	authUser := authcontext.GetUser(ctx)

	newItem := entity.Item{
		Name:      req.Name,
		Price:     &req.Price,
		IsActive:  types.SafelyReference(true),
		CreatedBy: &authUser.ID,
		UpdatedBy: &authUser.ID,
	}
	if req.Description != "" {
		newItem.Description = &req.Description
	}

	if err := i.itemRepo.Create(ctx, &newItem); err != nil {
		return entity.ItemResponse{}, errorLib.InternalServerError("")
	}

	i.activityLog.Log(context.WithoutCancel(ctx), entity.ActivityLogInsertRequest{
		ActivityEvent: logger.AuditLogEvent{
			Type:     []logger.AuditLogEventType{logger.CreationType},
			Category: []logger.AuditLogEventCategory{logger.DatabaseCategory},
			Action:   logger.SensitiveCreateEventAction,
			Outcome:  logger.SuccessEventOutcome,
		},
		ActivityName:     "Menambahkan item: " + req.Name,
		AdditionalFields: map[string]any{"itemId": newItem.ID},
	})

	return toItemResponse(newItem), nil
}

func (i *item) GetByID(ctx context.Context, req entity.ItemRequest) (entity.ItemResponse, error) {
	found, err := i.itemRepo.GetByID(ctx, req.ID)
	if err != nil {
		return entity.ItemResponse{}, err
	}

	return toItemResponse(found), nil
}

func (i *item) GetList(ctx context.Context, req entity.GetListItemRequest) ([]entity.ItemResponse, *entity.PaginationResponse, error) {
	searchColumnsByKey := map[string][]string{
		entity.ItemSearchByName: {"items.name"},
	}
	searchColumns, ok := searchColumnsByKey[req.SearchBy]
	if req.SearchBy != "" && !ok {
		return []entity.ItemResponse{}, entity.BuildPaginationResponse(req.PaginationRequest, 0, 0), nil
	}

	items, pagination, err := i.itemRepo.GetList(ctx, req.PaginationRequest, searchColumns)
	if err != nil {
		return nil, nil, errorLib.InternalServerError("")
	}

	result := make([]entity.ItemResponse, 0, len(items))
	for _, it := range items {
		result = append(result, toItemResponse(it))
	}
	return result, pagination, nil
}

func (i *item) Update(ctx context.Context, req entity.UpdateItemRequest) error {
	if err := i.validator.Bind(req, validation.ItemUpdate); err != nil {
		return err
	}

	authUser := authcontext.GetUser(ctx)

	updates := entity.Item{
		Name:      req.Name,
		Price:     &req.Price,
		IsActive:  &req.IsActive,
		UpdatedBy: &authUser.ID,
	}
	if req.Description != "" {
		updates.Description = &req.Description
	}

	if err := i.itemRepo.Update(ctx, req.ID, updates); err != nil {
		return err
	}

	i.activityLog.Log(context.WithoutCancel(ctx), entity.ActivityLogInsertRequest{
		ActivityEvent: logger.AuditLogEvent{
			Type:     []logger.AuditLogEventType{logger.ChangeType},
			Category: []logger.AuditLogEventCategory{logger.DatabaseCategory},
			Action:   logger.SensitiveUpdateEventAction,
			Outcome:  logger.SuccessEventOutcome,
		},
		ActivityName:     "Memperbarui item: " + req.Name,
		AdditionalFields: map[string]any{"itemId": req.ID},
	})

	return nil
}

func (i *item) Delete(ctx context.Context, req entity.ItemRequest) error {
	authUser := authcontext.GetUser(ctx)

	if err := i.itemRepo.Delete(ctx, req.ID, authUser.ID); err != nil {
		return err
	}

	i.activityLog.Log(context.WithoutCancel(ctx), entity.ActivityLogInsertRequest{
		ActivityEvent: logger.AuditLogEvent{
			Type:     []logger.AuditLogEventType{logger.DeletionType},
			Category: []logger.AuditLogEventCategory{logger.DatabaseCategory},
			Action:   logger.SensitiveDeleteEventAction,
			Outcome:  logger.SuccessEventOutcome,
		},
		ActivityName:     "Menghapus item",
		AdditionalFields: map[string]any{"itemId": req.ID},
	})

	return nil
}

func toItemResponse(it entity.Item) entity.ItemResponse {
	return entity.ItemResponse{
		ID:          it.ID,
		Name:        it.Name,
		Description: types.SafelyDereference(it.Description),
		Price:       types.SafelyDereference(it.Price),
		IsActive:    types.SafelyDereference(it.IsActive),
		CreatedAt:   it.CreatedAt,
		UpdatedAt:   it.UpdatedAt,
	}
}
