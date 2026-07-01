package usecase

import (
	"context"

	"github.com/yerobalg/wealthpulse-service/helper/authcontext"
	"github.com/yerobalg/wealthpulse-service/helper/cryptolib"
	errorLib "github.com/yerobalg/wealthpulse-service/helper/errors"
	"github.com/yerobalg/wealthpulse-service/helper/types"
	"github.com/yerobalg/wealthpulse-service/helper/validator"

	"github.com/yerobalg/wealthpulse-service/src/entity"
	"github.com/yerobalg/wealthpulse-service/src/entity/validation"
	"github.com/yerobalg/wealthpulse-service/src/repository"
)

type UserInterface interface {
	Login(ctx context.Context, userLoginRequest entity.UserLoginRequest) (entity.UserLoginResponse, error)
	GetProfile(ctx context.Context) authcontext.User
	CreateUser(ctx context.Context, req entity.CreateUserRequest) (entity.User, error)
	UpdateUser(ctx context.Context, req entity.UpdateUserRequest) error
	GetList(ctx context.Context, req entity.GetListUserRequest) ([]entity.UserListResponse, *entity.PaginationResponse, error)
	EnsureSuperuser(ctx context.Context, req entity.EnsureSuperuserRequest) error
}

type user struct {
	userRepo       repository.UserInterface
	roleRepo       repository.RoleInterface
	permissionRepo repository.PermissionInterface
	password       cryptolib.PasswordInterface
	jwt            cryptolib.JWTInterface
	validator      validator.Interface
}

type UserInitParam struct {
	UserRepo       repository.UserInterface
	RoleRepo       repository.RoleInterface
	PermissionRepo repository.PermissionInterface
	Password       cryptolib.PasswordInterface
	JWT            cryptolib.JWTInterface
	Validator      validator.Interface
}

func InitUser(param UserInitParam) UserInterface {
	return &user{
		userRepo:       param.UserRepo,
		roleRepo:       param.RoleRepo,
		permissionRepo: param.PermissionRepo,
		password:       param.Password,
		jwt:            param.JWT,
		validator:      param.Validator,
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

	return userResponse, nil
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

	return newUser, nil
}

// EnsureSuperuser idempotently seeds the single owner at startup from env. It
// inserts the user with the admin role only if the username does not already
// exist, storing the pre-hashed password verbatim (no re-hashing). No activity
// log is emitted: this runs at boot with no authenticated user or request context.
func (u *user) EnsureSuperuser(ctx context.Context, req entity.EnsureSuperuserRequest) error {
	if req.Username == "" || req.PasswordHash == "" {
		return nil
	}

	_, err := u.userRepo.Get(ctx, entity.UserRequest{Username: req.Username})
	if err == nil {
		return nil
	}
	if !errorLib.Is(err, errorLib.TypeNotFound) {
		return err
	}

	role, err := u.roleRepo.GetByCode(ctx, entity.RoleCodeAdmin)
	if err != nil {
		return err
	}

	newUser := entity.User{
		Username:           req.Username,
		Password:           req.PasswordHash,
		Name:               req.Name,
		IsMale:             &req.IsMale,
		RoleID:             role.ID,
		HasChangedPassword: types.SafelyReference(true),
		IsInactive:         types.SafelyReference(false),
	}

	return u.userRepo.Create(ctx, &newUser)
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

	return nil
}
