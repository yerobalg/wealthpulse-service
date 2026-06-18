package repository

import (
	"github.com/yerobalg/wealthpulse-service/helper/db"
)

type Repository struct {
	User         UserInterface
	Role         RoleInterface
	Permission   PermissionInterface
	RevokedToken RevokedTokenInterface
	ActivityLog  ActivityLogInterface
	Item         ItemInterface
}

func Init(dbConn db.DB) *Repository {
	return &Repository{
		User:         InitUser(dbConn),
		Role:         InitRole(dbConn),
		Permission:   InitPermission(dbConn),
		RevokedToken: InitRevokedToken(dbConn),
		ActivityLog:  InitActivityLog(dbConn),
		Item:         InitItem(dbConn),
	}
}
