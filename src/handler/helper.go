package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"github.com/yerobalg/wealthpulse-service/helper/appcontext"
	"github.com/yerobalg/wealthpulse-service/helper/errors"

	"github.com/yerobalg/wealthpulse-service/src/entity"
)

func (r *rest) BindParam(ctx *gin.Context, param any) error {
	if err := ctx.ShouldBindUri(param); err != nil {
		return err
	}

	return ctx.ShouldBindWith(param, binding.Query)
}

func (r *rest) BindBody(ctx *gin.Context, body any) error {
	return ctx.ShouldBindWith(body, binding.Default(ctx.Request.Method, ctx.ContentType()))
}

func (r *rest) SuccessResponse(ctx *gin.Context, message string, data any, pg *entity.PaginationResponse) {
	ctx.JSON(200, entity.HTTPResponse{
		Meta:       getRequestMetadata(ctx),
		Message:    entity.ResponseMessage{Title: "Sukses", Description: message},
		IsSuccess:  true,
		Data:       data,
		Pagination: pg,
	})
	r.log.Info(ctx.Request.Context(), message, nil)
}

func (r *rest) CreatedResponse(ctx *gin.Context, message string, data any) {
	ctx.JSON(201, entity.HTTPResponse{
		Meta: getRequestMetadata(ctx),
		Message: entity.ResponseMessage{
			Title:       "Sukses",
			Description: message,
		},
		IsSuccess: true,
		Data:      data,
	})
	r.log.Info(ctx.Request.Context(), message, data)
}

func (r *rest) ErrorResponse(ctx *gin.Context, err error) {
	ctx.JSON(int(errors.GetCode(err)), entity.HTTPResponse{
		Meta: getRequestMetadata(ctx),
		Message: entity.ResponseMessage{
			Title:       errors.GetType(err),
			Description: errors.GetMessage(err),
		},
		IsSuccess: false,
		Data:      errors.GetData(err),
	})
	if errors.IsApplicationError(err) {
		r.log.Error(ctx.Request.Context(), err.Error(), errors.GetAll(err))
	} else {
		r.log.Error(ctx.Request.Context(), err.Error())
	}
}

func getRequestMetadata(ctx *gin.Context) entity.Meta {
	meta := entity.Meta{
		RequestID: appcontext.GetRequestId(ctx.Request.Context()),
		Time:      time.Now().Format(time.RFC3339),
	}

	requestStartTime := appcontext.GetRequestStartTime(ctx.Request.Context())
	if !requestStartTime.IsZero() {
		elapsedTimeMs := time.Since(requestStartTime).Milliseconds()
		meta.TimeElapsed = fmt.Sprintf("%dms", elapsedTimeMs)
	}

	return meta
}
