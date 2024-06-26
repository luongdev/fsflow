package processors

import (
	"fmt"
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/session/input"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type TransferProcessor struct {
	*FreeswitchActivityProcessorImpl
}

func NewTransferProcessor(w shared.FreeswitchWorkflow, aP session.ActivityProvider) *TransferProcessor {
	return &TransferProcessor{FreeswitchActivityProcessorImpl: NewFreeswitchActivityProcessor(w, aP)}
}

func (p *TransferProcessor) Process(ctx workflow.Context, metadata shared.Metadata) (o *shared.WorkflowOutput, err error) {
	logger := workflow.GetLogger(ctx)
	o = shared.NewWorkflowOutput(metadata.GetSessionId())

	oi := input.TransferWorkflowInput{}
	if ok := shared.ConvertInput(metadata.GetInput(), &oi); !ok {
		logger.Error("Failed to get input")
		err = fmt.Errorf("cannot cast input to TransferWorkflowInput")
		return
	}

	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		ExecutionStartToCloseTimeout: oi.Timeout,
		ParentClosePolicy:            client.ParentClosePolicyTerminate,
	})

	err = workflow.ExecuteChildWorkflow(ctx, "workflows.TransferWorkflow", oi).Get(ctx, o)

	//err = workflow.ExecuteChildWorkflow(ctx, "workflows.OfferWorkflow", oi).Get(ctx, o)
	if err != nil {
		if cadence.IsTimeoutError(err) {

		}
		logger.Error("Failed to execute child workflow", zap.Error(err))
		return
	}

	return
}

var _ shared.FreeswitchActivityProcessor = (*TransferProcessor)(nil)
