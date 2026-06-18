package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yerobalg/wealthpulse-service/src/entity"
)

// @Summary Get List of Items
// @Description Get a paginated list of items. Search supports name.
// @Tags Item
// @Produce json
// @Param limit query int false "Page size (default 10, max 100)."
// @Param page query int false "Page number (default 1)."
// @Param search_by query string false "Field to search by." Enums(name)
// @Param search_value query string false "Search value (case-insensitive substring match)."
// @Success 200 {object} entity.HTTPResponse{data=[]entity.ItemResponse}
// @Failure 401 {object} entity.HTTPResponse{}
// @Failure 500 {object} entity.HTTPResponse{}
// @Security BearerAuth
// @Router /item [GET]
func (r *rest) GetListItem(c *gin.Context) {
	ctx := c.Request.Context()

	var req entity.GetListItemRequest
	if err := r.BindParam(c, &req); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	data, pagination, err := r.usecase.Item.GetList(ctx, req)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.SuccessResponse(c, "Data item berhasil diambil", data, pagination)
}

// @Summary Get Item Detail
// @Description Get a single item by ID
// @Tags Item
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} entity.HTTPResponse{data=entity.ItemResponse}
// @Failure 401 {object} entity.HTTPResponse{}
// @Failure 404 {object} entity.HTTPResponse{}
// @Security BearerAuth
// @Router /item/{id} [GET]
func (r *rest) GetItemDetail(c *gin.Context) {
	ctx := c.Request.Context()

	var req entity.ItemRequest
	if err := r.BindParam(c, &req); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	data, err := r.usecase.Item.GetByID(ctx, req)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.SuccessResponse(c, "Detail item berhasil diambil", data, nil)
}

// @Summary Create Item
// @Description Create a new item
// @Tags Item
// @Accept json
// @Produce json
// @Param createItemBody body entity.CreateItemRequest true "Create item request body"
// @Success 200 {object} entity.HTTPResponse{data=entity.ItemResponse}
// @Failure 400 {object} entity.HTTPResponse{}
// @Failure 401 {object} entity.HTTPResponse{}
// @Failure 500 {object} entity.HTTPResponse{}
// @Security BearerAuth
// @Router /item [POST]
func (r *rest) CreateItem(c *gin.Context) {
	ctx := c.Request.Context()

	var req entity.CreateItemRequest
	if err := r.BindBody(c, &req); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	data, err := r.usecase.Item.Create(ctx, req)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.SuccessResponse(c, "Item berhasil ditambahkan", data, nil)
}

// @Summary Update Item
// @Description Update an existing item
// @Tags Item
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Param updateItemBody body entity.UpdateItemRequest true "Update item request body"
// @Success 200 {object} entity.HTTPResponse{}
// @Failure 400 {object} entity.HTTPResponse{}
// @Failure 401 {object} entity.HTTPResponse{}
// @Failure 404 {object} entity.HTTPResponse{}
// @Failure 500 {object} entity.HTTPResponse{}
// @Security BearerAuth
// @Router /item/{id} [PUT]
func (r *rest) UpdateItem(c *gin.Context) {
	ctx := c.Request.Context()

	var req entity.UpdateItemRequest
	if err := r.BindParam(c, &req); err != nil {
		r.ErrorResponse(c, err)
		return
	}
	if err := r.BindBody(c, &req); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := r.usecase.Item.Update(ctx, req); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.SuccessResponse(c, "Item berhasil diperbarui", nil, nil)
}

// @Summary Delete Item
// @Description Soft-delete an existing item
// @Tags Item
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} entity.HTTPResponse{}
// @Failure 401 {object} entity.HTTPResponse{}
// @Failure 404 {object} entity.HTTPResponse{}
// @Failure 500 {object} entity.HTTPResponse{}
// @Security BearerAuth
// @Router /item/{id} [DELETE]
func (r *rest) DeleteItem(c *gin.Context) {
	ctx := c.Request.Context()

	var req entity.ItemRequest
	if err := r.BindParam(c, &req); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := r.usecase.Item.Delete(ctx, req); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.SuccessResponse(c, "Item berhasil dihapus", nil, nil)
}
