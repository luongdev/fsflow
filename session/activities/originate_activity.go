package activities

import (
	"context"
	"github.com/google/uuid"
	"github.com/luongdev/fsflow/errors"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
	"time"
)

type OriginateActivityInput struct {
	shared.WorkflowInput

	UId         string                 `json:"uid"`
	Timeout     time.Duration          `json:"timeout"`
	ANI         string                 `json:"ani"`
	DNIS        string                 `json:"dnis"`
	OrigFrom    string                 `json:"origFrom"`
	OrigTo      string                 `json:"origTo"`
	Gateway     string                 `json:"gateway"`
	Profile     string                 `json:"profile"`
	AutoAnswer  bool                   `json:"autoAnswer"`
	AllowReject bool                   `json:"allowReject"`
	Direction   shared.Direction       `json:"direction"`
	Variables   map[string]interface{} `json:"variables"`
	Extension   string                 `json:"extension"`
	Background  bool                   `json:"background"`
}

type OriginateActivity struct {
	p freeswitch.SocketProvider
}

const OriginateActivityName = "activities.OriginateActivity"

func (o *OriginateActivity) Name() string {
	return OriginateActivityName
}

func NewOriginateActivity(p freeswitch.SocketProvider) *OriginateActivity {
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

		if input.Variables == nil {
			input.Variables = make(map[string]interface{})
		}

		if input.ANI == "" {
			input.ANI = input.OrigFrom
		}
		input.Variables["X-ANI"] = input.ANI

		if input.DNIS == "" {
			input.DNIS = input.OrigTo
		}
		input.Variables["X-DNIS"] = input.DNIS

		res, err := client.Originate(ctx, &freeswitch.Originator{
			SessionId:   input.GetSessionId(),
			Timeout:     input.Timeout,
			ANI:         input.ANI,
			DNIS:        input.DNIS,
			OrigFrom:    input.OrigFrom,
			OrigTo:      input.OrigTo,
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
		if input.Background {
			output.Metadata[shared.FieldUniqueId] = res
		} else {
			output.Metadata[shared.FieldOutput] = res
		}

		return output, nil
	}
}

var _ shared.FreeswitchActivity = (*OriginateActivity)(nil)
