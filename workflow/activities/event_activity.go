package activities

import (
	"context"
	"fmt"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

type EventActivityInput struct {
	shared.WorkflowInput

	EventName     string                 `json:"eventName"`
	EventSubClass string                 `json:"eventSubClass"`
	Headers       map[string]interface{} `json:"headers"`
}

type EventActivity struct {
	p freeswitch.SocketProvider
}

func (c *EventActivity) Name() string {
	return "activities.EventActivity"
}

func NewEventActivity(p freeswitch.SocketProvider) *EventActivity {
	return &EventActivity{p: p}
}

func (c *EventActivity) Handler() shared.ActivityFunc {
	return func(ctx context.Context, i shared.WorkflowInput) (*shared.WorkflowOutput, error) {
		logger := activity.GetLogger(ctx)
		output := shared.NewWorkflowOutput(i.GetSessionId())

		if err := i.Validate(); err != nil {
			logger.Error("Invalid input", zap.Any("input", i), zap.Error(err))
			return output, err
		}

		client := c.p.GetClient(i.GetSessionId())

		input := EventActivityInput{}
		ok := shared.ConvertInput(i, &input)

		if !ok {
			logger.Error("Failed to cast input to EventActivityInput")
			return output, shared.NewWorkflowInputError("Cannot cast input to EventActivityInput")
		}

		if input.EventName == "" {
			input.EventName = "CUSTOM"
		}

		if input.EventSubClass == "" {
			input.EventSubClass = "callmanager::event"
		}

		res, err := client.Execute(ctx, &freeswitch.Command{
			Uid:     input.GetSessionId(),
			AppName: "Event",
			AppArgs: fmt.Sprintf("Event-Name=%s,Event-Subclass=%s", input.EventName, input.EventSubClass),
		})

		if err != nil {
			return output, err
		}

		output.Success = true
		output.Metadata[shared.FieldMessage] = res

		return output, nil
	}
}

var _ shared.FreeswitchActivity = (*EventActivity)(nil)
