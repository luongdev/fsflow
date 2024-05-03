package activities

import (
	"context"
	"github.com/google/uuid"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
	"time"
)

type OriginateActivityInput struct {
	Timeout      time.Duration          `json:"timeout"`
	DialedNumber string                 `json:"dialedNumber"`
	Destination  string                 `json:"destination"`
	Gateway      string                 `json:"gateway"`
	Profile      string                 `json:"profile"`
	AutoAnswer   bool                   `json:"autoAnswer"`
	AllowReject  bool                   `json:"allowReject"`
	Direction    freeswitch.Direction   `json:"direction"`
	Variables    map[string]interface{} `json:"variables"`
	BridgeTo     string                 `json:"bridgeTo"`
}

type OriginateActivity struct {
	fsClient *freeswitch.SocketClient
}

func (o *OriginateActivity) Name() string {
	return "activities.OriginateActivity"
}

func NewOriginateActivity(fsClient *freeswitch.SocketClient) *OriginateActivity {
	return &OriginateActivity{fsClient: fsClient}
}

func (o *OriginateActivity) Handler() shared.ActivityFunc {
	return func(ctx context.Context, i shared.WorkflowInput) (*shared.WorkflowOutput, error) {
		logger := activity.GetLogger(ctx)
		output := shared.NewWorkflowOutput(i.GetSessionId())

		if err := i.Validate(); err != nil {
			logger.Error("Invalid input", zap.Any("input", i), zap.Error(err))
			return output, err
		}

		input := OriginateActivityInput{}
		ok := shared.ConvertInput(i, &input)

		info := activity.GetInfo(ctx)
		logger.Info("Executing OriginateActivity", zap.Any("info", info))

		if !ok {
			logger.Error("Failed to cast input to OriginateActivityInput")
			return output, shared.NewWorkflowInputError("Cannot cast input to OriginateActivityInput")
		}

		res, err := (*o.fsClient).Originate(ctx, &freeswitch.Originator{
			Timeout:     input.Timeout,
			ANI:         input.DialedNumber,
			DNIS:        input.Destination,
			Direction:   input.Direction,
			Profile:     input.Profile,
			Gateway:     input.Gateway,
			AutoAnswer:  input.AutoAnswer,
			AllowReject: input.AllowReject,
			Variables:   input.Variables,
		})
		if err != nil {
			return output, err
		}

		output.Success = true
		output.Metadata[shared.FieldSessionId] = res

		var id uuid.UUID
		if input.BridgeTo == "" && ctx.Value(shared.FieldSessionId) != nil {
			id, err = uuid.Parse(ctx.Value(shared.FieldSessionId).(string))
		} else {
			id, err = uuid.Parse(input.BridgeTo)
		}
		if err == nil {
			output.Metadata[shared.FieldAction] = shared.ActionBridge
			output.Metadata[shared.FieldInput] = BridgeActivityInput{Originator: id.String(), Originatee: res}
		}

		return output, nil
	}
}

var _ shared.FreeswitchActivity = (*OriginateActivity)(nil)
