package activities

import (
	"context"
	"fmt"
	"github.com/luongdev/fsflow/errors"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

type SetActivityInput struct {
	UId       string                 `json:"uid"`
	Variables map[string]interface{} `json:"variables"`
}

type SetActivity struct {
	p freeswitch.SocketProvider
}

const SetActivityName = "activities.SetActivity"

func (c *SetActivity) Name() string {
	return SetActivityName
}

func NewSetActivity(p freeswitch.SocketProvider) *SetActivity {
	return &SetActivity{p: p}
}

func (c *SetActivity) Handler() shared.ActivityFunc {
	return func(ctx context.Context, i shared.WorkflowInput) (*shared.WorkflowOutput, error) {
		logger := activity.GetLogger(ctx)
		output := shared.NewWorkflowOutput(i.GetSessionId())

		if err := i.Validate(); err != nil {
			logger.Error("Invalid input", zap.Any("input", i), zap.Error(err))
			return output, err
		}

		client := c.p.GetClient(i.GetSessionId())

		input := SetActivityInput{}
		ok := shared.ConvertInput(i, &input)
		if !ok {
			logger.Error("Failed to cast input to SetActivityInput")
			return output, errors.NewWorkflowInputError("Cannot cast input to SetActivityInput")
		}

		cmd := "set"
		args := ""
		del := ";"
		multi := false
		if len(input.Variables) == 0 {
			return output, nil
		} else if len(input.Variables) > 1 {
			multi = true
			cmd = "multiset"
			args = fmt.Sprintf("^^")
		}

		for k, v := range input.Variables {
			arg := fmt.Sprintf("%s=%v", k, v)
			if multi {
				args += fmt.Sprintf("%s%s", del, arg)
				continue
			}
			args = arg
		}

		res, err := client.Execute(ctx, &freeswitch.Command{AppName: cmd, AppArgs: args, Uid: input.UId})
		if err != nil {
			return output, err
		}

		output.Success = true
		output.Metadata[shared.FieldMessage] = res

		logger.Info("SetActivity completed", zap.Any("input", input))

		return output, nil
	}
}

var _ shared.FreeswitchActivity = (*SetActivity)(nil)
