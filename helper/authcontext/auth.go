package authcontext

import (
	"context"
)

type key string

const (
	userAuthInfo key = "UserAuthInfo"
)

type Role struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Permission struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type User struct {
	ID                 int64        `json:"id"`
	Username           string       `json:"username"`
	Name               string       `json:"name"`
	Gender             string       `json:"gender"`
	HasChangedPassword bool         `json:"hasChangedPassword"`
	Role               Role         `json:"role"`
	Permissions        []Permission `json:"permissions"`
	UserToken          string       `json:"-"`
}

// GetUserID retrieves the authenticated user's ID from the context. Returns 0 if not set.
func GetUserID(ctx context.Context) int64 {
	user, ok := ctx.Value(userAuthInfo).(User)
	if !ok {
		return 0
	}

	return user.ID
}

// GetUser retrieves the authenticated User from the context. Returns an empty User if not set.
func GetUser(ctx context.Context) User {
	user, ok := ctx.Value(userAuthInfo).(User)
	if !ok {
		return User{}
	}

	return user
}

// SetUserDirect stores the given User directly into the context.
func SetUserDirect(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userAuthInfo, user)
}

// SetUser parses the JWT claims map and stores the authenticated User (including role and permissions) into the context.
func SetUser(ctx context.Context, user map[string]any, userToken string) context.Context {
	var role Role
	if r, ok := user["role"].(map[string]any); ok {
		role = Role{
			Name: r["name"].(string),
			Code: r["code"].(string),
		}
	}

	var permissions []Permission
	if rawPerms, ok := user["permissions"].([]any); ok {
		for _, p := range rawPerms {
			if pm, ok := p.(map[string]any); ok {
				permissions = append(permissions, Permission{
					Name: pm["name"].(string),
					Code: pm["code"].(string),
				})
			}
		}
	}

	gender, _ := user["gender"].(string)

	userObj := User{
		ID:                 int64(user["id"].(float64)),
		Username:           user["username"].(string),
		Name:               user["name"].(string),
		Gender:             gender,
		HasChangedPassword: user["hasChangedPassword"].(bool),
		Role:               role,
		Permissions:        permissions,
		UserToken:          userToken,
	}
	return context.WithValue(ctx, userAuthInfo, userObj)
}
