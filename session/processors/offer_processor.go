package processors

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/session/activities"
	input2 "github.com/luongdev/fsflow/session/input"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence"
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

func (p *OfferProcessor) Process(ctx workflow.Context, metadata shared.Metadata) (*shared.WorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	output := shared.NewWorkflowOutput(metadata.GetSessionId())

	input := input2.OfferWorkflowInput{}
	if ok := shared.ConvertInput(metadata.GetInput(), &input); !ok {
		logger.Error("Failed to get input")
		return output, fmt.Errorf("cannot cast input to OfferWorkflowInput")
	}

	info := workflow.GetInfo(ctx)
	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		ExecutionStartToCloseTimeout: input.Timeout,
		WorkflowID:                   info.WorkflowExecution.RunID,
	})

	if input.Variables == nil {
		input.Variables = make(map[string]interface{})
	}

	offerId, err := uuid.NewRandom()
	if err != nil {
		logger.Error("Failed to generate UUID", zap.Error(err))
		return output, err
	}

	logger.Info("Offering call", zap.String("uuid", offerId.String()))

	input.Variables["uuid"] = offerId.String()
	err = workflow.ExecuteChildWorkflow(ctx, "workflows.OfferWorkflow", input).Get(ctx, &output)
	if err != nil {
		if cadence.IsTimeoutError(err) {
			hA := p.aP.GetActivity(activities.HangupActivityName)
			if hA == nil {
				err = fmt.Errorf("activity %s not found", shared.ActionHangup)
				return output, err
			}
			cCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				ScheduleToStartTimeout: time.Second,
				StartToCloseTimeout:    5 * time.Second,
			})
			err = workflow.ExecuteActivity(cCtx, hA.Handler(), activities.HangupActivityInput{
				SessionId:    offerId.String(),
				HangupCause:  "NORMAL_CLEARING",
				HangupReason: "OfferTimeout",
			}).Get(cCtx, output)
		}
		logger.Error("Failed to execute child workflow", zap.Error(err))
		return output, err
	}

	return output, err
}

var _ shared.FreeswitchActivityProcessor = (*OfferProcessor)(nil)
