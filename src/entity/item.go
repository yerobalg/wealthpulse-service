package entity

import (
	"gorm.io/gorm"
)

// Item is a sample business resource that demonstrates the full layered
// convention (entity -> repository -> usecase -> handler). Use it as the
// reference when scaffolding a new resource, then delete it.
type Item struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy" gorm:"default:null"`
	UpdatedBy *int64         `json:"updatedBy" gorm:"default:null"`
	DeletedBy *int64         `json:"deletedBy" gorm:"default:null"`

	Name        string  `json:"name" gorm:"not null;type:varchar(255)"`
	Description *string `json:"description" gorm:"type:text;default:null"`
	// Price is a pointer because zero is a valid, meaningful value — GORM must be
	// able to distinguish "not set" from "intentionally zero" (see CLAUDE.md).
	Price    *int64 `json:"price" gorm:"not null;default:0"`
	IsActive *bool  `json:"isActive" gorm:"not null;default:true"`
}

const (
	ItemSearchByName = "name"
)

type ItemRequest struct {
	ID int64 `uri:"id" param:"id"`
}

type GetListItemRequest struct {
	PaginationRequest
}

type CreateItemRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Price       int64  `json:"price" validate:"gte=0"`
}

// UpdateItemRequest combines the path param (id) and the body fields in a
// single struct — bind the param first, then the body (see CLAUDE.md).
type UpdateItemRequest struct {
	ID          int64  `uri:"id" json:"-" validate:"required,gt=0"`
	Name        string `json:"name" validate:"required,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Price       int64  `json:"price" validate:"gte=0"`
	IsActive    bool   `json:"isActive"`
}

type ItemResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	IsActive    bool   `json:"isActive"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}
