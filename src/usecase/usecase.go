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
	User UserInterface
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
	userParam := UserInitParam{
		UserRepo:       param.Repository.User,
		RoleRepo:       param.Repository.Role,
		PermissionRepo: param.Repository.Permission,
		Password:       param.Password,
		JWT:            param.JWT,
		Validator:      validator.Init(),
	}

	return &Usecase{
		User: InitUser(userParam),
	}
}
