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
	SessionId   string `json:"sessionId"`
	HangupCause string `json:"hangupCause"`
}

type HangupActivity struct {
	fsClient *freeswitch.SocketClient
}

func (c *HangupActivity) Name() string {
	return "activities.HangupActivity"
}

func NewHangupActivity(fsClient *freeswitch.SocketClient) *HangupActivity {
	return &HangupActivity{fsClient: fsClient}
}

func (c *HangupActivity) Handler() shared.ActivityFunc {
	return func(ctx context.Context, i interface{}) (*shared.WorkflowOutput, error) {
		logger := activity.GetLogger(ctx)
		output := &shared.WorkflowOutput{Success: false, Metadata: make(shared.Metadata)}
		input := HangupActivityInput{}
		ok := shared.Convert(i, &input)

		if !ok {
			logger.Error("Failed to cast input to HangupActivityInput")
			return output, fmt.Errorf("failed to cast input to HangupActivityInput")
		}

		_, err := (*c.fsClient).Api(ctx, &freeswitch.Command{
			AppName: "uuid_kill",
			AppArgs: fmt.Sprintf("%v %v", input.SessionId, input.HangupCause),
		})

		if err != nil {
			logger.Error("Failed to execute command", zap.Error(err))
			return output, err
		}

		output.Success = true
		output.Metadata[shared.FieldSessionId] = input.SessionId
		output.Metadata[shared.FieldMessage] =
			fmt.Sprintf("Session %v has been hungup cause: %v", input.SessionId, input.HangupCause)

		return output, nil
	}
}

var _ shared.FreeswitchActivity = (*HangupActivity)(nil)
