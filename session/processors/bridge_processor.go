package processors

import (
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/session/activities"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type BridgeProcessor struct {
	*FreeswitchActivityProcessorImpl
}

func NewBridgeProcessor(w shared.FreeswitchWorkflow, aP session.ActivityProvider) *BridgeProcessor {
	return &BridgeProcessor{FreeswitchActivityProcessorImpl: NewFreeswitchActivityProcessor(w, aP)}
}

func (p *BridgeProcessor) Process(ctx workflow.Context, metadata shared.Metadata) (*shared.WorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	output := shared.NewWorkflowOutput(metadata.GetSessionId())

	i := activities.BridgeActivityInput{}
	err := p.GetInput(metadata, &i)
	if err != nil {
		logger.Error("Failed to get input", zap.Error(err))
		return output, err
	}

	if i.TransferLeg != "" {
		hA := p.aP.GetActivity(activities.HangupActivityName)
		if hA != nil {
			err = workflow.ExecuteActivity(ctx, hA.Handler(), &activities.HangupActivityInput{
				WorkflowInput: i.WorkflowInput,
				UId:           i.TransferLeg,
				HangupReason:  "ManualTransfer",
				HangupCause:   "NORMAL_CLEARING",
			}).Get(ctx, &output)
			if err != nil {
				logger.Error("Failed to execute hangup activity", zap.Error(err))
			}
		}
	}

	bA := p.aP.GetActivity(activities.BridgeActivityName)
	err = workflow.ExecuteActivity(ctx, bA.Handler(), i).Get(ctx, &output)

	return output, err
}

var _ shared.FreeswitchActivityProcessor = (*BridgeProcessor)(nil)
