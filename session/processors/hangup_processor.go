package processors

import (
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/session/activities"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type HangupProcessor struct {
	*FreeswitchActivityProcessorImpl
}

func NewHangupProcessor(w shared.FreeswitchWorkflow, aP session.ActivityProvider) *HangupProcessor {
	return &HangupProcessor{FreeswitchActivityProcessorImpl: NewFreeswitchActivityProcessor(w, aP)}
}

func (p *HangupProcessor) Process(ctx workflow.Context, metadata shared.Metadata) (*shared.WorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	output := shared.NewWorkflowOutput(metadata.GetSessionId())

	i := activities.HangupActivityInput{}
	err := p.GetInput(metadata, &i)
	if err != nil {
		logger.Error("Failed to get input", zap.Error(err))
		return output, err
	}

	hA := p.aP.GetActivity(activities.HangupActivityName)
	err = workflow.ExecuteActivity(ctx, hA.Handler(), i).Get(ctx, &output)

	return output, err
}

var _ shared.FreeswitchActivityProcessor = (*BridgeProcessor)(nil)
