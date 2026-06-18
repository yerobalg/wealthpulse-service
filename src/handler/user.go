package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/yerobalg/wealthpulse-service/src/entity"
)

// @Summary Login
// @Description Login for user
// @Tags User
// @Accept json
// @Produce json
// @Param loginBody body entity.UserLoginRequest true "Login request body"
// @Success 200 {object} entity.HTTPResponse{data=entity.UserLoginResponse}
// @Failure 400 {object} entity.HTTPResponse{}
// @Failure 401 {object} entity.HTTPResponse{}
// @Failure 500 {object} entity.HTTPResponse{}
// @Router /user/login [POST]
func (r *rest) Login(c *gin.Context) {
	ctx := c.Request.Context()

	var loginRequest entity.UserLoginRequest
	if err := r.BindBody(c, &loginRequest); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	res, err := r.usecase.User.Login(ctx, loginRequest)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.SuccessResponse(c, "Login berhasil", res, nil)
}

// @Summary Logout
// @Description Logout and revoke the current user token
// @Tags User
// @Produce json
// @Success 200 {object} entity.HTTPResponse{}
// @Failure 401 {object} entity.HTTPResponse{}
// @Failure 500 {object} entity.HTTPResponse{}
// @Security BearerAuth
// @Router /user/logout [POST]
func (r *rest) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	if err := r.usecase.User.Logout(ctx); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.SuccessResponse(c, "Logout berhasil", nil, nil)
}

// @Summary Get Profile
// @Description Get the authenticated user's profile from token
// @Tags User
// @Produce json
// @Success 200 {object} entity.HTTPResponse{data=authcontext.User}
// @Failure 401 {object} entity.HTTPResponse{}
// @Security BearerAuth
// @Router /user/profile [GET]
func (r *rest) GetProfile(c *gin.Context) {
	ctx := c.Request.Context()
	res := r.usecase.User.GetProfile(ctx)
	r.SuccessResponse(c, "Profil berhasil diambil", res, nil)
}

// @Summary Change Password
// @Description Change password for authenticated user
// @Tags User
// @Accept json
// @Produce json
// @Param changePasswordBody body entity.ChangePasswordRequest true "Change password request body"
// @Success 200 {object} entity.HTTPResponse{}
// @Failure 400 {object} entity.HTTPResponse{}
// @Failure 401 {object} entity.HTTPResponse{}
// @Failure 500 {object} entity.HTTPResponse{}
// @Security BearerAuth
// @Router /user/password [PATCH]
func (r *rest) ChangePassword(c *gin.Context) {
	ctx := c.Request.Context()

	var changePasswordRequest entity.ChangePasswordRequest
	if err := r.BindBody(c, &changePasswordRequest); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	err := r.usecase.User.ChangePassword(ctx, changePasswordRequest)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.SuccessResponse(c, "Password berhasil diubah", nil, nil)
}

// @Summary Update User
// @Description Update an existing user account
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param updateUserBody body entity.UpdateUserRequest true "Update user request body"
// @Success 200 {object} entity.HTTPResponse{}
// @Failure 400 {object} entity.HTTPResponse{}
// @Failure 401 {object} entity.HTTPResponse{}
// @Failure 404 {object} entity.HTTPResponse{}
// @Failure 500 {object} entity.HTTPResponse{}
// @Security BearerAuth
// @Router /user/{id} [PUT]
func (r *rest) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()

	var req entity.UpdateUserRequest
	if err := r.BindParam(c, &req); err != nil {
		r.ErrorResponse(c, err)
		return
	}
	if err := r.BindBody(c, &req); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	if err := r.usecase.User.UpdateUser(ctx, req); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.SuccessResponse(c, "Pengguna berhasil diperbarui", nil, nil)
}

// @Summary Get List of Users
// @Description Get a paginated list of users with their role detail. Search supports name, username, or both.
// @Tags User
// @Produce json
// @Param limit query int false "Page size (default 10, max 100)."
// @Param page query int false "Page number (default 1)."
// @Param search_by query string false "Field to search by (name, username, name,username)."
// @Param search_value query string false "Search value (case-insensitive substring match)."
// @Success 200 {object} entity.HTTPResponse{data=[]entity.UserListResponse}
// @Failure 400 {object} entity.HTTPResponse{}
// @Failure 401 {object} entity.HTTPResponse{}
// @Failure 500 {object} entity.HTTPResponse{}
// @Security BearerAuth
// @Router /user [GET]
func (r *rest) GetListUser(c *gin.Context) {
	ctx := c.Request.Context()

	var req entity.GetListUserRequest
	if err := r.BindParam(c, &req); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	data, pagination, err := r.usecase.User.GetList(ctx, req)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.SuccessResponse(c, "Data user berhasil diambil", data, pagination)
}

// @Summary Create User
// @Description Create a new user account (assigned the default "user" role)
// @Tags User
// @Accept json
// @Produce json
// @Param createUserBody body entity.CreateUserRequest true "Create user request body"
// @Success 200 {object} entity.HTTPResponse{data=entity.User}
// @Failure 400 {object} entity.HTTPResponse{}
// @Failure 401 {object} entity.HTTPResponse{}
// @Failure 404 {object} entity.HTTPResponse{}
// @Failure 500 {object} entity.HTTPResponse{}
// @Security BearerAuth
// @Router /user [POST]
func (r *rest) CreateUser(c *gin.Context) {
	ctx := c.Request.Context()

	var req entity.CreateUserRequest
	if err := r.BindBody(c, &req); err != nil {
		r.ErrorResponse(c, err)
		return
	}

	res, err := r.usecase.User.CreateUser(ctx, req)
	if err != nil {
		r.ErrorResponse(c, err)
		return
	}

	r.SuccessResponse(c, "Pengguna berhasil ditambahkan", res, nil)
}
