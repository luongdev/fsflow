package activities

import (
	"context"
	"fmt"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

type BridgeActivityInput struct {
	Originator string `json:"originator"`
	Originatee string `json:"originatee"`

	shared.WorkflowInput
}

type BridgeActivity struct {
	fsClient *freeswitch.SocketClient
}

func (c *BridgeActivity) Name() string {
	return "activities.BridgeActivity"
}

func NewBridgeActivity(fsClient *freeswitch.SocketClient) *BridgeActivity {
	return &BridgeActivity{fsClient: fsClient}
}

func (c *BridgeActivity) Handler() shared.ActivityFunc {
	return func(ctx context.Context, i shared.WorkflowInput) (*shared.WorkflowOutput, error) {
		logger := activity.GetLogger(ctx)
		output := shared.NewWorkflowOutput(i.GetSessionId())

		if err := i.Validate(); err != nil {
			logger.Error("Invalid input", zap.Any("input", i), zap.Error(err))
			return output, err
		}

		input := BridgeActivityInput{}
		ok := shared.ConvertInput(i, &input)

		if !ok {
			logger.Error("Failed to cast input to BridgeActivityInput")
			return output, shared.NewWorkflowInputError("Cannot cast input to BridgeActivityInput")
		}

		res, err := (*c.fsClient).Api(ctx, &freeswitch.Command{
			AppName: "uuid_bridge",
			AppArgs: fmt.Sprintf("%v %v", input.Originator, input.Originatee),
		})

		if err != nil {
			return output, err
		}

		output.Success = true
		output.Metadata[shared.FieldMessage] = res

		logger.Info("BridgeActivity completed", zap.Any("input", input))

		return output, nil
	}
}

var _ shared.FreeswitchActivity = (*BridgeActivity)(nil)
