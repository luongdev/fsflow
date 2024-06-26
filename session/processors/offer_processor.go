package processors

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/session/activities"
	"github.com/luongdev/fsflow/session/input"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"time"
)

type OfferProcessor struct {
	*FreeswitchActivityProcessorImpl
}

func NewOfferProcessor(w shared.FreeswitchWorkflow, aP session.ActivityProvider) *OfferProcessor {
	return &OfferProcessor{FreeswitchActivityProcessorImpl: NewFreeswitchActivityProcessor(w, aP)}
}

func (p *OfferProcessor) Process(ctx workflow.Context, metadata shared.Metadata) (output *shared.WorkflowOutput, err error) {
	logger := workflow.GetLogger(ctx)
	output = shared.NewWorkflowOutput(metadata.GetSessionId())

	oi := input.OfferWorkflowInput{}
	if ok := shared.ConvertInput(metadata.GetInput(), &oi); !ok {
		logger.Error("Failed to get input")
		err = fmt.Errorf("cannot cast input to OfferWorkflowInput")
		return
	}

	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		ExecutionStartToCloseTimeout: oi.Timeout,
		ParentClosePolicy:            client.ParentClosePolicyTerminate,
	})

	oi.UId, err = uuid.NewRandom()
	if err != nil {
		logger.Error("Failed to generate UUID", zap.Error(err))
		return
	}

	err = workflow.ExecuteChildWorkflow(ctx, "workflows.OfferWorkflow", oi).Get(ctx, &output)
	if err != nil {
		if cadence.IsTimeoutError(err) {
			hA := p.aP.GetActivity(activities.HangupActivityName)
			if hA == nil {
				err = fmt.Errorf("activity %s not found", shared.ActionHangup)
				return
			}
			cCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				ScheduleToStartTimeout: time.Second,
				StartToCloseTimeout:    5 * time.Second,
			})
			err = workflow.ExecuteActivity(cCtx, hA.Handler(), activities.HangupActivityInput{
				UId:          oi.UId.String(),
				HangupCause:  "ORIGINATOR_CANCEL",
				HangupReason: "OfferTimeout",
			}).Get(cCtx, output)
		}
		logger.Error("Failed to execute child workflow", zap.Error(err))
		return
	}

	return
}

var _ shared.FreeswitchActivityProcessor = (*OfferProcessor)(nil)
