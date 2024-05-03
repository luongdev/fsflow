package processors

import (
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"github.com/luongdev/fsflow/workflow/activities"
	libworkflow "go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type OriginateProcessor struct {
	*FreeswitchActivityProcessorImpl
}

func NewOriginateProcessor(client *freeswitch.SocketClient) *OriginateProcessor {
	return &OriginateProcessor{FreeswitchActivityProcessorImpl: NewFreeswitchActivityProcessor(client)}
}

func (p *OriginateProcessor) Process(ctx libworkflow.Context, metadata shared.Metadata) (*shared.WorkflowOutput, error) {
	logger := libworkflow.GetLogger(ctx)
	output := &shared.WorkflowOutput{Success: false, Metadata: make(shared.Metadata)}

	i := activities.OriginateActivityInput{}
	err := p.GetInput(metadata, &i)
	if err != nil {
		logger.Error("Failed to get input", zap.Error(err))
		return output, err
	}

	originateActivity := activities.NewOriginateActivity(p.FsClient)
	err = libworkflow.ExecuteActivity(ctx, originateActivity.Handler(), i).Get(ctx, &output)

	return output, err
}

var _ shared.FreeswitchActivityProcessor = (*OriginateProcessor)(nil)
