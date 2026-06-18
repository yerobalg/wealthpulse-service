package usecase

import (
	"context"

	"github.com/yerobalg/wealthpulse-service/helper/authcontext"

	"github.com/yerobalg/wealthpulse-service/src/entity"
	"github.com/yerobalg/wealthpulse-service/src/repository"
)

type RevokedTokenInterface interface {
	IsTokenRevoked(ctx context.Context, token string) (bool, error)
	RevokeToken(ctx context.Context, reason entity.RevokedTokenReason, expiredAt *int64) error
}

type revokedToken struct {
	revokedTokenRepo repository.RevokedTokenInterface
}

func InitRevokedToken(revokedTokenRepo repository.RevokedTokenInterface) RevokedTokenInterface {
	return &revokedToken{revokedTokenRepo: revokedTokenRepo}
}

func (r *revokedToken) IsTokenRevoked(ctx context.Context, token string) (bool, error) {
	return r.revokedTokenRepo.IsTokenRevoked(ctx, token)
}

func (r *revokedToken) RevokeToken(ctx context.Context, reason entity.RevokedTokenReason, expiredAt *int64) error {
	user := authcontext.GetUser(ctx)

	revokedToken := entity.RevokedToken{
		UserID:    user.ID,
		Token:     user.UserToken,
		Reason:    reason,
		ExpiredAt: expiredAt,
		CreatedBy: &user.ID,
		UpdatedBy: &user.ID,
	}

	return r.revokedTokenRepo.Create(ctx, revokedToken)
}
