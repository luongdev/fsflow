package processors

import (
	"context"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"github.com/luongdev/fsflow/workflow/activities"
	libworkflow "go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type OriginateProcessor struct {
	*FreeswitchActivityProcessorImpl
}

func NewOriginateProcessor(client freeswitch.SocketProvider) *OriginateProcessor {
	return &OriginateProcessor{FreeswitchActivityProcessorImpl: NewFreeswitchActivityProcessor(client)}
}

func (p *OriginateProcessor) Process(ctx libworkflow.Context, metadata shared.Metadata) (*shared.WorkflowOutput, error) {
	logger := libworkflow.GetLogger(ctx)
	output := shared.NewWorkflowOutput(metadata.GetSessionId())

	i := activities.OriginateActivityInput{}
	err := p.GetInput(metadata, &i)
	if err != nil {
		logger.Error("Failed to get input", zap.Error(err))
		return output, err
	}

	client := p.SocketProvider.GetClient(i.GetSessionId())

	if i.Background {
		bCtx, cancel := context.WithTimeout(context.Background(), i.Timeout)
		err := client.AllEvents(bCtx)
		err = client.MyEvents(bCtx, i.GetSessionId())
		if err != nil {
			logger.Error("Failed to add filter", zap.Error(err))
			cancel()
		}

		defer cancel()
	}

	ctx = libworkflow.WithActivityOptions(ctx, libworkflow.ActivityOptions{
		StartToCloseTimeout:    i.Timeout,
		ScheduleToStartTimeout: 1,
	})

	originateActivity := activities.NewOriginateActivity(p.SocketProvider)
	err = libworkflow.ExecuteActivity(ctx, originateActivity.Handler(), i).Get(ctx, &output)

	return output, err
}

var _ shared.FreeswitchActivityProcessor = (*OriginateProcessor)(nil)
