package usecase

import (
	"context"

	"github.com/yerobalg/wealthpulse-service/helper/authcontext"
	"github.com/yerobalg/wealthpulse-service/helper/cryptolib"
	errorLib "github.com/yerobalg/wealthpulse-service/helper/errors"
	"github.com/yerobalg/wealthpulse-service/helper/logger"
	"github.com/yerobalg/wealthpulse-service/helper/types"
	"github.com/yerobalg/wealthpulse-service/helper/validator"

	"github.com/yerobalg/wealthpulse-service/src/entity"
	"github.com/yerobalg/wealthpulse-service/src/entity/validation"
	"github.com/yerobalg/wealthpulse-service/src/repository"
)

type UserInterface interface {
	Login(ctx context.Context, userLoginRequest entity.UserLoginRequest) (entity.UserLoginResponse, error)
	ChangePassword(ctx context.Context, changePasswordRequest entity.ChangePasswordRequest) error
	Logout(ctx context.Context) error
	GetProfile(ctx context.Context) authcontext.User
	CreateUser(ctx context.Context, req entity.CreateUserRequest) (entity.User, error)
	UpdateUser(ctx context.Context, req entity.UpdateUserRequest) error
	GetList(ctx context.Context, req entity.GetListUserRequest) ([]entity.UserListResponse, *entity.PaginationResponse, error)
}

type user struct {
	userRepo         repository.UserInterface
	roleRepo         repository.RoleInterface
	permissionRepo   repository.PermissionInterface
	revokedTokenRepo repository.RevokedTokenInterface
	password         cryptolib.PasswordInterface
	jwt              cryptolib.JWTInterface
	activityLog      ActivityLogInterface
	validator        validator.Interface
	txManager        TransactionManager
}

type UserInitParam struct {
	UserRepo         repository.UserInterface
	RoleRepo         repository.RoleInterface
	PermissionRepo   repository.PermissionInterface
	RevokedTokenRepo repository.RevokedTokenInterface
	Password         cryptolib.PasswordInterface
	JWT              cryptolib.JWTInterface
	ActivityLog      ActivityLogInterface
	Validator        validator.Interface
	TxManager        TransactionManager
}

func InitUser(param UserInitParam) UserInterface {
	return &user{
		userRepo:         param.UserRepo,
		roleRepo:         param.RoleRepo,
		permissionRepo:   param.PermissionRepo,
		revokedTokenRepo: param.RevokedTokenRepo,
		password:         param.Password,
		jwt:              param.JWT,
		activityLog:      param.ActivityLog,
		validator:        param.Validator,
		txManager:        param.TxManager,
	}
}

const (
	UsernameOrPasswordWrong = "Username tidak ditemukan atau password salah"
)

func (u *user) Login(ctx context.Context, userLoginRequest entity.UserLoginRequest) (entity.UserLoginResponse, error) {
	var userResponse entity.UserLoginResponse

	if err := u.validator.Bind(userLoginRequest, validation.UserLogin); err != nil {
		return userResponse, err
	}

	userRequest := entity.UserRequest{
		Username: userLoginRequest.Username,
	}

	user, err := u.userRepo.Get(ctx, userRequest)
	if errorLib.Is(err, errorLib.TypeNotFound) {
		return userResponse, errorLib.Unauthorized(UsernameOrPasswordWrong)
	} else if err != nil {
		return userResponse, err
	}

	if !u.password.Compare(user.Password, userLoginRequest.Password) {
		return userResponse, errorLib.Unauthorized(UsernameOrPasswordWrong)
	}

	userPermissions, err := u.permissionRepo.GetByRoleID(ctx, user.RoleID)
	if err != nil {
		return userResponse, err
	}

	permissionResponses := make([]entity.PermissionResponse, len(userPermissions))
	for i, p := range userPermissions {
		permissionResponses[i] = entity.PermissionResponse{
			Name: p.Name,
			Code: p.Code,
		}
	}

	token, err := u.jwt.Encode(user.ToJWTClaims(permissionResponses))
	if err != nil {
		return userResponse, err
	}

	userResponse = entity.UserLoginResponse{
		User:        user,
		Permissions: permissionResponses,
		AccessToken: token,
	}

	logCtx := authcontext.SetUserDirect(context.WithoutCancel(ctx), authcontext.User{
		ID:        user.ID,
		UserToken: token,
	})
	u.activityLog.Log(logCtx, entity.ActivityLogInsertRequest{
		ActivityEvent: logger.AuditLogEvent{
			Type:     []logger.AuditLogEventType{logger.StartType},
			Category: []logger.AuditLogEventCategory{logger.AuthCategory},
			Action:   logger.LoginSuccessEventAction,
			Outcome:  logger.SuccessEventOutcome,
		},
		ActivityName: "Login",
	})

	return userResponse, nil
}

func (u *user) ChangePassword(ctx context.Context, changePasswordRequest entity.ChangePasswordRequest) error {
	if err := u.validator.Bind(changePasswordRequest, validation.Password("newPassword", "Password baru")); err != nil {
		return err
	}

	authUser := authcontext.GetUser(ctx)

	userQuery := entity.UserRequest{
		ID: authUser.ID,
	}

	user, err := u.userRepo.Get(ctx, userQuery)
	if errorLib.Is(err, errorLib.TypeNotFound) {
		return errorLib.Unauthorized(UsernameOrPasswordWrong)
	} else if err != nil {
		return err
	}

	newPasswordHash, err := u.password.Hash(changePasswordRequest.NewPassword)
	if err != nil {
		return err
	}

	user.Password = newPasswordHash

	logActivityName := "Mengganti password"
	if !types.SafelyDereference(user.HasChangedPassword) {
		logActivityName += " untuk pertama kali"
	}
	user.HasChangedPassword = types.SafelyReference(true)

	if err := u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := u.userRepo.Update(txCtx, userQuery, user); err != nil {
			return err
		}

		return u.revokedTokenRepo.Create(txCtx, entity.RevokedToken{
			UserID: authUser.ID,
			Token:  authUser.UserToken,
			Reason: entity.RevokedTokenReasonPasswordChanged,
		})
	}); err != nil {
		return err
	}

	u.activityLog.Log(context.WithoutCancel(ctx), entity.ActivityLogInsertRequest{
		ActivityEvent: logger.AuditLogEvent{
			Type:     []logger.AuditLogEventType{logger.ChangeType},
			Category: []logger.AuditLogEventCategory{logger.AuthCategory},
			Action:   logger.PasswordChangeEventAction,
			Outcome:  logger.SuccessEventOutcome,
		},
		ActivityName: logActivityName,
	})

	return nil
}

func (u *user) GetProfile(ctx context.Context) authcontext.User {
	return authcontext.GetUser(ctx)
}

func (u *user) CreateUser(ctx context.Context, req entity.CreateUserRequest) (entity.User, error) {
	if err := u.validator.Bind(req, validation.UserCreate); err != nil {
		return entity.User{}, err
	}

	authUser := authcontext.GetUser(ctx)

	role, err := u.roleRepo.GetByCode(ctx, entity.RoleCodeUser)
	if err != nil {
		return entity.User{}, err
	}

	passwordHash, err := u.password.Hash(req.Password)
	if err != nil {
		return entity.User{}, err
	}

	newUser := entity.User{
		Username:  req.Username,
		Password:  passwordHash,
		Name:      req.Name,
		IsMale:    &req.IsMale,
		RoleID:    role.ID,
		CreatedBy: &authUser.ID,
		UpdatedBy: &authUser.ID,
	}

	if err := u.userRepo.Create(ctx, &newUser); err != nil {
		return entity.User{}, err
	}

	u.activityLog.Log(context.WithoutCancel(ctx), entity.ActivityLogInsertRequest{
		ActivityEvent: logger.AuditLogEvent{
			Type:     []logger.AuditLogEventType{logger.CreationType},
			Category: []logger.AuditLogEventCategory{logger.IAMCategory},
			Action:   logger.SensitiveCreateEventAction,
			Outcome:  logger.SuccessEventOutcome,
		},
		ActivityName: "Menambahkan pengguna: " + req.Username,
	})

	return newUser, nil
}

func (u *user) GetList(ctx context.Context, req entity.GetListUserRequest) ([]entity.UserListResponse, *entity.PaginationResponse, error) {
	searchColumnsByKey := map[string][]string{
		entity.UserSearchByName:         {"users.name"},
		entity.UserSearchByUsername:     {"users.username"},
		entity.UserSearchByNameUsername: {"users.name", "users.username"},
	}
	searchColumns, ok := searchColumnsByKey[req.SearchBy]
	if req.SearchBy != "" && !ok {
		return []entity.UserListResponse{}, entity.BuildPaginationResponse(req.PaginationRequest, 0, 0), nil
	}

	users, pagination, err := u.userRepo.GetListWithRole(ctx, req.PaginationRequest, searchColumns)
	if err != nil {
		return nil, nil, errorLib.InternalServerError("")
	}

	result := make([]entity.UserListResponse, 0, len(users))
	for _, usr := range users {
		result = append(result, toUserListResponse(usr))
	}
	return result, pagination, nil
}

func toUserListResponse(usr entity.User) entity.UserListResponse {
	return entity.UserListResponse{
		ID:                 usr.ID,
		Username:           usr.Username,
		Name:               usr.Name,
		Position:           usr.Position,
		IsMale:             types.SafelyDereference(usr.IsMale),
		HasChangedPassword: types.SafelyDereference(usr.HasChangedPassword),
		IsInactive:         types.SafelyDereference(usr.IsInactive),
		Role: entity.UserRoleResponse{
			ID:   usr.Role.ID,
			Name: usr.Role.Name,
			Code: usr.Role.Code,
		},
		CreatedAt: usr.CreatedAt,
		UpdatedAt: usr.UpdatedAt,
	}
}

func (u *user) UpdateUser(ctx context.Context, req entity.UpdateUserRequest) error {
	if err := u.validator.Bind(req, validation.UserUpdate); err != nil {
		return err
	}

	role, err := u.roleRepo.GetByCode(ctx, entity.RoleCodeUser)
	if err != nil {
		return err
	}

	authUser := authcontext.GetUser(ctx)
	updates := entity.User{
		Username:   req.Username,
		Name:       req.Name,
		IsMale:     &req.IsMale,
		IsInactive: &req.IsInactive,
		UpdatedBy:  &authUser.ID,
	}

	err = u.userRepo.Update(ctx, entity.UserRequest{
		ID:     req.ID,
		RoleID: role.ID,
	}, updates)
	if errorLib.Is(err, errorLib.TypeNotFound) {
		return errorLib.NotFound("Pengguna")
	}
	if err != nil {
		return err
	}

	u.activityLog.Log(context.WithoutCancel(ctx), entity.ActivityLogInsertRequest{
		ActivityEvent: logger.AuditLogEvent{
			Type:     []logger.AuditLogEventType{logger.ChangeType},
			Category: []logger.AuditLogEventCategory{logger.IAMCategory},
			Action:   logger.SensitiveUpdateEventAction,
			Outcome:  logger.SuccessEventOutcome,
		},
		ActivityName: "Memperbarui pengguna: " + req.Username,
		AdditionalFields: map[string]any{
			"userId":     req.ID,
			"isInactive": req.IsInactive,
		},
	})

	return nil
}

func (u *user) Logout(ctx context.Context) error {
	authUser := authcontext.GetUser(ctx)

	return u.revokedTokenRepo.Create(ctx, entity.RevokedToken{
		UserID:    authUser.ID,
		Token:     authUser.UserToken,
		Reason:    entity.RevokedTokenReasonLogout,
		CreatedBy: &authUser.ID,
		UpdatedBy: &authUser.ID,
	})
}
