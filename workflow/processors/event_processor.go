package processors

import (
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"github.com/luongdev/fsflow/workflow/activities"
	libworkflow "go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type EventProcessor struct {
	*FreeswitchActivityProcessorImpl
}

func NewEventProcessor(client freeswitch.SocketProvider) *EventProcessor {
	return &EventProcessor{FreeswitchActivityProcessorImpl: NewFreeswitchActivityProcessor(client)}
}

func (p *EventProcessor) Process(ctx libworkflow.Context, metadata shared.Metadata) (*shared.WorkflowOutput, error) {
	logger := libworkflow.GetLogger(ctx)
	output := shared.NewWorkflowOutput(metadata.GetSessionId())

	i := activities.EventActivityInput{}
	err := p.GetInput(metadata, &i)
	if err != nil {
		logger.Error("Failed to get input", zap.Error(err))
		return output, err
	}

	eventActivity := activities.NewEventActivity(p.SocketProvider)
	err = libworkflow.ExecuteActivity(ctx, eventActivity.Handler(), i).Get(ctx, &output)

	return output, err
}

var _ shared.FreeswitchActivityProcessor = (*EventProcessor)(nil)
