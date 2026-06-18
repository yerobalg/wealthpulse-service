package usecase

import (
	"context"
	"encoding/json"

	"github.com/yerobalg/wealthpulse-service/helper/appcontext"
	"github.com/yerobalg/wealthpulse-service/helper/async"
	"github.com/yerobalg/wealthpulse-service/helper/authcontext"
	"github.com/yerobalg/wealthpulse-service/helper/logger"
	"github.com/yerobalg/wealthpulse-service/helper/types"

	"github.com/yerobalg/wealthpulse-service/src/entity"
	"github.com/yerobalg/wealthpulse-service/src/repository"
)

type ActivityLogInterface interface {
	Log(ctx context.Context, req entity.ActivityLogInsertRequest)
}

type activityLog struct {
	activityLogRepo repository.ActivityLogInterface
	asyncLib        async.Interface
	log             logger.Interface
}

func InitActivityLog(activityLogRepo repository.ActivityLogInterface, asyncLib async.Interface, log logger.Interface) ActivityLogInterface {
	return &activityLog{activityLogRepo: activityLogRepo, asyncLib: asyncLib, log: log}
}

func (a *activityLog) Log(ctx context.Context, req entity.ActivityLogInsertRequest) {
	a.asyncLib.Run(ctx, func() {
		user := authcontext.GetUser(ctx)

		metadataJSON, err := json.Marshal(appcontext.GetMetadata(ctx))
		if err != nil {
			a.log.Error(ctx, "failed to marshal metadata", err.Error())
			return
		}

		activityEventJSON, err := json.Marshal(req.ActivityEvent)
		if err != nil {
			a.log.Error(ctx, "failed to marshal activity event", err.Error())
			return
		}

		activityLog := entity.ActivityLog{
			UserID:        user.ID,
			UserToken:     user.UserToken,
			Metadata:      string(metadataJSON),
			ActivityEvent: string(activityEventJSON),
			ActivityName:  req.ActivityName,
		}

		if req.AdditionalFields != nil {
			additionalFieldsJSON, err := json.Marshal(req.AdditionalFields)
			if err != nil {
				a.log.Error(ctx, "failed to marshal additional fields", err.Error())
				return
			}
			additionalFields := string(additionalFieldsJSON)
			activityLog.AdditionalFields = types.SafelyReference(additionalFields)
		}

		if err := a.activityLogRepo.Create(ctx, activityLog); err != nil {
			a.log.Error(ctx, "failed to create activity log", err.Error())
			return
		}

		a.log.Info(ctx, "activity log created",
			"userToken", user.UserToken,
			"activityEvent", string(activityEventJSON),
			"activityName", req.ActivityName,
			"additionalFields", activityLog.AdditionalFields,
		)
	})
}
