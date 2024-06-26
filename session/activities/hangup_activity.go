package activities

import (
	"context"
	"fmt"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

type HangupActivityInput struct {
	UId          string `json:"uid"`
	HangupCause  string `json:"hangupCause"`
	HangupReason string `json:"hangupReason"`

	shared.WorkflowInput
}

type HangupActivity struct {
	p freeswitch.SocketProvider
}

const HangupActivityName = "activities.HangupActivity"

func (c *HangupActivity) Name() string {
	return HangupActivityName
}

func NewHangupActivity(p freeswitch.SocketProvider) *HangupActivity {
	return &HangupActivity{p: p}
}

func (c *HangupActivity) Handler() shared.ActivityFunc {
	return func(ctx context.Context, i shared.WorkflowInput) (*shared.WorkflowOutput, error) {
		logger := activity.GetLogger(ctx)
		output := shared.NewWorkflowOutput(i.GetSessionId())

		if err := i.Validate(); err != nil {
			logger.Error("Invalid input", zap.Any("input", i), zap.Error(err))
			return output, err
		}

		client := c.p.GetClient(i.GetSessionId())

		input := HangupActivityInput{}
		if !shared.ConvertInput(i, &input) {
			logger.Error("Failed to cast input to HangupActivityInput")
			return output, fmt.Errorf("failed to cast input to HangupActivityInput")
		}

		if input.UId == "" {
			input.UId = i.GetSessionId()
		}

		if input.HangupReason != "" {
			res, err := client.Execute(ctx, &freeswitch.Command{
				Uid:     input.UId,
				AppName: "set",
				AppArgs: fmt.Sprintf("hangup_reason %v", input.HangupReason),
			})

			if err != nil {
				logger.Error("Failed to set hangup reason", zap.Error(err), zap.Any("response", res))
			}
		}

		_, err := client.Api(ctx, &freeswitch.Command{
			AppName: "uuid_kill",
			AppArgs: fmt.Sprintf("%v %v", input.UId, input.HangupCause),
		})

		if err != nil {
			logger.Error("Failed to execute command", zap.Error(err))
			return output, err
		}

		output.Success = true
		output.Metadata[shared.FieldSessionId] = input.UId
		output.Metadata[shared.FieldMessage] =
			fmt.Sprintf("Session %v has been hungup cause: %v", input.UId, input.HangupCause)

		return output, nil
	}
}

var _ shared.FreeswitchActivity = (*HangupActivity)(nil)
