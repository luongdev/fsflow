package activities

import (
	"context"
	"fmt"
	"github.com/luongdev/fsflow/errors"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/session/input"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

type GetActivityInput struct {
	UId      string `json:"uid"`
	Variable string `json:"variable"`

	input.WorkflowInput
}

type GetActivity struct {
	p freeswitch.SocketProvider
}

const GetActivityName = "activities.GetActivity"

func (c *GetActivity) Name() string {
	return GetActivityName
}

func NewGetActivity(p freeswitch.SocketProvider) *GetActivity {
	return &GetActivity{p: p}
}

func (c *GetActivity) Handler() shared.ActivityFunc {
	return func(ctx context.Context, i shared.WorkflowInput) (*shared.WorkflowOutput, error) {
		logger := activity.GetLogger(ctx)
		output := shared.NewWorkflowOutput(i.GetSessionId())

		if err := i.Validate(); err != nil {
			logger.Error("Invalid gi", zap.Any("gi", i), zap.Error(err))
			return output, err
		}

		client := c.p.GetClient(i.GetSessionId())
		gi := GetActivityInput{}
		ok := shared.ConvertInput(i, &gi)
		if !ok {
			logger.Error("Failed to cast gi to GetActivityInput")
			return output, errors.NewWorkflowInputError("Cannot cast gi to GetActivityInput")
		}

		args := fmt.Sprintf("%v %v", i.GetSessionId(), gi.Variable)
		if gi.UId != "" {
			args = fmt.Sprintf("%v %v", gi.UId, gi.Variable)
		}

		res, err := client.Api(ctx, &freeswitch.Command{AppName: "uuid_getvar", AppArgs: args})
		if err != nil {
			return output, err
		}

		if res == "_undef_" {
			res = ""
		}

		output.Success = true
		output.Metadata[shared.FieldUniqueId] = res
		output.Metadata[shared.FieldOutput] = res

		logger.Info("GetActivity completed", zap.Any("gi", gi))

		return output, nil
	}
}

var _ shared.FreeswitchActivity = (*GetActivity)(nil)
