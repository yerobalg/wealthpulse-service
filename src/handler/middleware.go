package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/yerobalg/wealthpulse-service/helper/appcontext"
	"github.com/yerobalg/wealthpulse-service/helper/authcontext"
	"github.com/yerobalg/wealthpulse-service/helper/errors"
	"github.com/yerobalg/wealthpulse-service/helper/types"

	"github.com/yerobalg/wealthpulse-service/src/entity"
)

// timeout middleware wraps the request context with a timeout
func (r *rest) SetTimeout(ctx *gin.Context) {
	// wrap the request context with a timeout, this will cause the request to fail if it takes more than defined timeout
	c, cancel := context.WithTimeout(ctx.Request.Context(), 5*time.Minute) // TODO: change this hardcoded timeout to config later

	// cancel to clear resources after finished
	defer cancel()
	c = appcontext.SetRequestStartTime(c, time.Now())

	// replace request with context wrapped request
	ctx.Request = ctx.Request.WithContext(c)
	ctx.Next()
}

func (r *rest) AddFieldsToContext(ctx *gin.Context) {
	requestID := uuid.New().String()

	c := ctx.Request.Context()
	c = appcontext.SetRequestId(c, requestID)
	c = appcontext.SetUserAgent(c, ctx.Request.Header.Get(appcontext.HeaderUserAgent))
	c = appcontext.SetDeviceType(c, ctx.Request.Header.Get(appcontext.HeaderDeviceType))
	ctx.Request = ctx.Request.WithContext(c)

	ctx.Next()
}

func (r *rest) CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func (r *rest) Authorization() gin.HandlerFunc {
	return r.checkToken
}

func (r *rest) checkToken(ctx *gin.Context) {
	header := ctx.Request.Header.Get("Authorization")
	if header == "" {
		r.ErrorResponse(ctx, errors.Unauthorized("Harap login terlebih dahulu"))
		ctx.Abort()
		return
	}

	jwtToken := header[len("Bearer "):]
	tokenClaims, err := r.jwt.Decode(jwtToken)
	if err != nil {
		if errors.GetMessage(err) == "Token expired" {
			bgCtx := context.WithoutCancel(ctx.Request.Context())
			r.async.Run(bgCtx, func() {
				r.revokeExpiredToken(bgCtx, tokenClaims, jwtToken)
			})
		}
		r.ErrorResponse(ctx, errors.Unauthorized("Token tidak valid"))
		ctx.Abort()
		return
	}

	data, ok := tokenClaims["data"].(map[string]any)
	if !ok {
		r.ErrorResponse(ctx, errors.Unauthorized("Token tidak valid"))
		ctx.Abort()
		return
	}

	isRevoked, err := r.usecase.RevokedToken.IsTokenRevoked(ctx.Request.Context(), jwtToken)
	if err != nil || isRevoked {
		r.ErrorResponse(ctx, errors.Unauthorized("Token tidak valid"))
		ctx.Abort()
		return
	}

	c := ctx.Request.Context()
	c = authcontext.SetUser(c, data, jwtToken)
	ctx.Request = ctx.Request.WithContext(c)

	ctx.Next()
}

func (r *rest) revokeExpiredToken(ctx context.Context, tokenClaims map[string]any, jwtToken string) {
	data, ok := tokenClaims["data"].(map[string]any)
	if !ok {
		return
	}

	id, ok := data["id"].(float64)
	if !ok {
		return
	}

	var expiredAt *int64
	if exp, ok := tokenClaims["exp"].(float64); ok {
		expiredAt = types.SafelyReference(int64(exp))
	}

	c := authcontext.SetUserDirect(ctx, authcontext.User{
		ID:        int64(id),
		UserToken: jwtToken,
	})
	r.usecase.RevokedToken.RevokeToken(c, entity.RevokedTokenReasonExpired, expiredAt)
}

func (r *rest) AuthorizeRole(roleCode string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userRole := authcontext.GetUser(ctx.Request.Context()).Role
		if userRole.Code != roleCode {
			r.ErrorResponse(ctx, errors.Unauthorized("Anda tidak memiliki akses untuk melakukan aksi ini"))
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func (r *rest) AuthorizePermission(permissionCode string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		permissions := authcontext.GetUser(ctx.Request.Context()).Permissions
		for _, p := range permissions {
			if p.Code == permissionCode {
				ctx.Next()
				return
			}
		}

		r.ErrorResponse(ctx, errors.Unauthorized("Anda tidak memiliki akses untuk melakukan aksi ini"))
		ctx.Abort()
	}
}
