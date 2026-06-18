package entity

import (
	"gorm.io/gorm"

	"github.com/yerobalg/wealthpulse-service/helper/types"
)

type User struct {
	ID        int64          `gorm:"primary_key" json:"id"`
	CreatedAt int64          `json:"createdAt"`
	UpdatedAt int64          `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedBy *int64         `json:"createdBy" gorm:"default:null"`
	UpdatedBy *int64         `json:"updatedBy" gorm:"default:null"`
	DeletedBy *int64         `json:"deletedBy" gorm:"default:null"`

	Username           string  `json:"username" gorm:"not null;unique;type:varchar(255)"`
	Password           string  `json:"-" gorm:"not null;type:text"`
	Name               string  `json:"name" gorm:"not null;type:varchar(255)"`
	Position           *string `json:"position" gorm:"type:varchar(255);default:null"`
	IsMale             *bool   `json:"isMale" gorm:"not null"`
	RoleID             int64   `json:"roleId" gorm:"index;not null"`
	HasChangedPassword *bool   `json:"hasChangedPassword" gorm:"type:bool;default:false"`
	IsInactive         *bool   `json:"isInactive" gorm:"not null;default:false"`

	Role Role `json:"role" gorm:"foreignKey:RoleID"`
}

type UserRequest struct {
	ID                 int64  `uri:"id" param:"id"`
	Username           string `json:"-" param:"username"`
	HasChangedPassword bool   `json:"-" param:"has_changed_password"`
	RoleID             int64  `json:"-" param:"role_id"`
	PaginationRequest
}

const (
	UserSearchByName         = "name"
	UserSearchByUsername     = "username"
	UserSearchByNameUsername = "name,username"
)

type GetListUserRequest struct {
	PaginationRequest
}

type UserRoleResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type UserListResponse struct {
	ID                 int64            `json:"id"`
	Username           string           `json:"username"`
	Name               string           `json:"name"`
	Position           *string          `json:"position"`
	IsMale             bool             `json:"isMale"`
	HasChangedPassword bool             `json:"hasChangedPassword"`
	IsInactive         bool             `json:"isInactive"`
	Role               UserRoleResponse `json:"role"`
	CreatedAt          int64            `json:"createdAt"`
	UpdatedAt          int64            `json:"updatedAt"`
}

type UserWithRoleAndPermissions struct {
	User        User
	Role        Role
	Permissions []Permission
}

type UserLoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type UserLoginResponse struct {
	User        User                 `json:"user"`
	Permissions []PermissionResponse `json:"permissions"`
	AccessToken string               `json:"accessToken"`
}

type ChangePasswordRequest struct {
	NewPassword string `json:"newPassword" validate:"required,min=8,printascii,containsany=0123456789,containsany=abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"`
}

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,max=255"`
	IsMale   bool   `json:"isMale"`
	Username string `json:"username" validate:"required,max=255,printascii"`
	Password string `json:"password" validate:"required,min=8,printascii,containsany=0123456789,containsany=abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"`
}

type UpdateUserRequest struct {
	ID         int64  `uri:"id" json:"-" validate:"required,gt=0"`
	Name       string `json:"name" validate:"required,max=255"`
	IsMale     bool   `json:"isMale"`
	Username   string `json:"username" validate:"required,max=255,printascii"`
	IsInactive bool   `json:"isInactive"`
}

func (u User) ToJWTClaims(permissions []PermissionResponse) map[string]any {
	userPerms := make([]map[string]any, len(permissions))
	for i, p := range permissions {
		userPerms[i] = map[string]any{
			"name": p.Name,
			"code": p.Code,
		}
	}

	gender := "Perempuan"
	if types.SafelyDereference(u.IsMale) {
		gender = "Laki-Laki"
	}

	return map[string]any{
		"id":                 u.ID,
		"username":           u.Username,
		"name":               u.Name,
		"gender":             gender,
		"hasChangedPassword": u.HasChangedPassword,
		"role": map[string]any{
			"name": u.Role.Name,
			"code": u.Role.Code,
		},
		"permissions": userPerms,
	}
}
