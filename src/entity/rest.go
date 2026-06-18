package entity

import (
	"strings"

	"gorm.io/gorm"
)

const (
	DefaultPaginationLimit = int64(10)
	MaxPaginationLimit     = int64(100)
	DefaultPaginationPage  = int64(1)
)

// ApplyPagination applies WHERE (search), LIMIT, and OFFSET to a query.
// Pass an empty/nil slice to skip the search filter. When multiple columns are provided,
// the WHERE clause ORs ILIKE across all of them. Columns must be table-qualified when the
// query joins another table (e.g. "groups.name") to avoid ambiguous-column errors.
func ApplyPagination(tx *gorm.DB, req PaginationRequest, searchColumns []string) *gorm.DB {
	tx = ApplySearchOnly(tx, req, searchColumns)

	limit := req.Limit
	if limit <= 0 {
		limit = DefaultPaginationLimit
	}
	if limit > MaxPaginationLimit {
		limit = MaxPaginationLimit
	}

	page := req.Page
	if page <= 0 {
		page = DefaultPaginationPage
	}

	return tx.Limit(int(limit)).Offset(int((page - 1) * limit))
}

// ApplySearchOnly applies only the WHERE filter — used by count queries that must share
// the same conditions as the data query but must not inherit LIMIT/OFFSET.
func ApplySearchOnly(tx *gorm.DB, req PaginationRequest, searchColumns []string) *gorm.DB {
	if len(searchColumns) == 0 || req.SearchValue == "" {
		return tx
	}

	pattern := "%" + req.SearchValue + "%"
	if len(searchColumns) == 1 {
		return tx.Where(searchColumns[0]+" ILIKE ?", pattern)
	}

	conds := make([]string, 0, len(searchColumns))
	args := make([]any, 0, len(searchColumns))
	for _, col := range searchColumns {
		conds = append(conds, col+" ILIKE ?")
		args = append(args, pattern)
	}
	return tx.Where(strings.Join(conds, " OR "), args...)
}

// BuildPaginationResponse constructs a PaginationResponse from a request, the current page element count,
// and the total element count across all pages.
func BuildPaginationResponse(req PaginationRequest, currentElements, totalElements int64) *PaginationResponse {
	limit := req.Limit
	if limit <= 0 {
		limit = DefaultPaginationLimit
	}
	if limit > MaxPaginationLimit {
		limit = MaxPaginationLimit
	}

	page := req.Page
	if page <= 0 {
		page = DefaultPaginationPage
	}

	totalPage := totalElements / limit
	if totalElements%limit != 0 {
		totalPage++
	}

	sortBy := req.SortBy
	if sortBy == nil {
		sortBy = []string{}
	}

	return &PaginationResponse{
		CurrentPage:    page,
		CurrentElement: currentElements,
		TotalPage:      totalPage,
		TotalElement:   totalElements,
		SortBy:         sortBy,
	}
}

type HTTPResponse struct {
	Meta       Meta                `json:"metaData"`
	Message    ResponseMessage     `json:"message"`
	IsSuccess  bool                `json:"isSuccess"`
	Data       any                 `json:"data"`
	Pagination *PaginationResponse `json:"pagination"`
}

type ResponseMessage struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Meta struct {
	Time        string `json:"timestamp"`
	RequestID   string `json:"requestId"`
	TimeElapsed string `json:"timeElapsed"`
}

type PaginationRequest struct {
	GroupBy     []string `db:"-"`
	SortBy      []string `db:"sort_by"`
	SearchBy    string   `form:"search_by" db:"search_by"`
	SearchValue string   `form:"search_value" db:"search_value"`
	Limit       int64    `form:"limit" db:"limit"`
	Page        int64    `form:"page" db:"page"`
}

type PaginationResponse struct {
	CurrentPage    int64    `json:"currentPage"`
	CurrentElement int64    `json:"currentElements"`
	TotalPage      int64    `json:"totalPages"`
	TotalElement   int64    `json:"totalElements"`
	SortBy         []string `json:"sortBy"`
}
