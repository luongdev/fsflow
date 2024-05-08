package activities

import (
	"context"
	"github.com/google/uuid"
	"github.com/luongdev/fsflow/errors"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/provider"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
	"time"
)

type OriginateActivityInput struct {
	shared.WorkflowInput

	Timeout      time.Duration          `json:"timeout"`
	DialedNumber string                 `json:"dialedNumber"`
	Destination  string                 `json:"destination"`
	Gateway      string                 `json:"gateway"`
	Profile      string                 `json:"profile"`
	AutoAnswer   bool                   `json:"autoAnswer"`
	AllowReject  bool                   `json:"allowReject"`
	Direction    freeswitch.Direction   `json:"direction"`
	Variables    map[string]interface{} `json:"variables"`
	Extension    string                 `json:"extension"`
	Background   bool                   `json:"background"`
	Callback     string                 `json:"callback"`
}

type OriginateActivity struct {
	p provider.SocketProvider
}

func (o *OriginateActivity) Name() string {
	return "activities.OriginateActivity"
}

func NewOriginateActivity(p provider.SocketProvider) *OriginateActivity {
	return &OriginateActivity{p: p}
}

func (o *OriginateActivity) Handler() shared.ActivityFunc {
	return func(ctx context.Context, i shared.WorkflowInput) (*shared.WorkflowOutput, error) {
		logger := activity.GetLogger(ctx)
		output := shared.NewWorkflowOutput(i.GetSessionId())

		client := o.p.GetClient(i.GetSessionId())

		input := OriginateActivityInput{}
		ok := shared.ConvertInput(i, &input)

		if !ok {
			logger.Error("Failed to cast input to OriginateActivityInput")
			return output, errors.NewWorkflowInputError("Cannot cast input to OriginateActivityInput")
		}

		if input.GetSessionId() == "" {
			s, err := uuid.NewRandom()
			if err == nil {
				input.WorkflowInput[shared.FieldSessionId] = s.String()
			}
		}

		res, err := client.Originate(ctx, &freeswitch.Originator{
			SessionId:   input.GetSessionId(),
			Callback:    input.Callback,
			Timeout:     input.Timeout,
			ANI:         input.DialedNumber,
			DNIS:        input.Destination,
			Direction:   input.Direction,
			Profile:     input.Profile,
			Gateway:     input.Gateway,
			AutoAnswer:  input.AutoAnswer,
			AllowReject: input.AllowReject,
			Variables:   input.Variables,
			Extension:   input.Extension,
			Background:  input.Background,
		})
		if err != nil {
			logger.Error("Failed to originate call", zap.Error(err))
			return output, nil
		}

		output.Success = true
		output.Metadata[shared.FieldUniqueId] = res

		return output, nil
	}
}

var _ shared.FreeswitchActivity = (*OriginateActivity)(nil)
