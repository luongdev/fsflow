package processors

import (
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/session/activities"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type CallbackProcessor struct {
	*FreeswitchActivityProcessorImpl
}

func NewCallbackProcessor(w shared.FreeswitchWorkflow, aP session.ActivityProvider) *CallbackProcessor {
	return &CallbackProcessor{FreeswitchActivityProcessorImpl: NewFreeswitchActivityProcessor(w, aP)}
}

func (p *CallbackProcessor) Process(ctx workflow.Context, metadata shared.Metadata) (*shared.WorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	output := shared.NewWorkflowOutput(metadata.GetSessionId())

	i := activities.CallbackActivityInput{}
	err := p.GetInput(metadata, &i)
	if err != nil {
		logger.Error("Failed to get input", zap.Error(err))
		return output, err
	}

	cA := p.aP.GetActivity(activities.CallbackActivityName)
	err = workflow.ExecuteActivity(ctx, cA.Handler(), i).Get(ctx, &output)

	return output, err
}

var _ shared.FreeswitchActivityProcessor = (*CallbackProcessor)(nil)
