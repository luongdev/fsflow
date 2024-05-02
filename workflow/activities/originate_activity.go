package activities

import (
	"context"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/activity"
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
	return func(ctx context.Context, i interface{}) (shared.WorkflowOutput, error) {
		logger := activity.GetLogger(ctx)
		output := shared.WorkflowOutput{Success: false, Metadata: make(shared.Metadata)}
		input := OriginateActivityInput{}
		ok := shared.Convert(i, &input)

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
		output.Metadata[shared.Uid] = res

		return output, nil
	}
}

var _ shared.FreeswitchActivity = (*OriginateActivity)(nil)