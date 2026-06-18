package usecase

import (
	"context"

	"github.com/yerobalg/wealthpulse-service/helper/async"
	"github.com/yerobalg/wealthpulse-service/helper/cryptolib"
	"github.com/yerobalg/wealthpulse-service/helper/logger"
	"github.com/yerobalg/wealthpulse-service/helper/validator"

	"github.com/yerobalg/wealthpulse-service/src/repository"
)

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

type Usecase struct {
	User         UserInterface
	ActivityLog  ActivityLogInterface
	RevokedToken RevokedTokenInterface
	Item         ItemInterface
}

type InitParam struct {
	Repository *repository.Repository
	Password   cryptolib.PasswordInterface
	JWT        cryptolib.JWTInterface
	Async      async.Interface
	Log        logger.Interface
	TxManager  TransactionManager
}

func Init(param InitParam) *Usecase {
	activityLog := InitActivityLog(param.Repository.ActivityLog, param.Async, param.Log)
	validator := validator.Init()

	userParam := UserInitParam{
		UserRepo:         param.Repository.User,
		RoleRepo:         param.Repository.Role,
		PermissionRepo:   param.Repository.Permission,
		RevokedTokenRepo: param.Repository.RevokedToken,
		Password:         param.Password,
		JWT:              param.JWT,
		ActivityLog:      activityLog,
		Validator:        validator,
		TxManager:        param.TxManager,
	}

	return &Usecase{
		User:         InitUser(userParam),
		ActivityLog:  activityLog,
		RevokedToken: InitRevokedToken(param.Repository.RevokedToken),
		Item: InitItem(ItemInitParam{
			ItemRepo:    param.Repository.Item,
			Validator:   validator,
			ActivityLog: activityLog,
		}),
	}
}
