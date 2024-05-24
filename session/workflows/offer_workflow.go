package workflows

import (
	"fmt"
	"github.com/google/uuid"
	fserrors "github.com/luongdev/fsflow/errors"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/session/activities"
	"github.com/luongdev/fsflow/session/input"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"time"
)

type OfferWorkflow struct {
	sP freeswitch.SocketProvider
	aP session.ActivityProvider

	r shared.WorkflowQueryResult
	e error
}

func (w *OfferWorkflow) Handler() shared.WorkflowFunc {
	retryIgnored := map[string]bool{
		"NO_ANSWER":        true,
		"NO_USER_RESPONSE": true,
		"USER_BUSY":        true,
		"NORMAL_CLEARING":  true,
	}
	return func(ctx workflow.Context, i shared.WorkflowInput) (o *shared.WorkflowOutput, err error) {
		logger := workflow.GetLogger(ctx)

		o = shared.NewWorkflowOutput(i.GetSessionId())
		wi := input.OfferWorkflowInput{}

		if ok := shared.ConvertInput(i, &wi); !ok {
			err = fserrors.NewWorkflowInputError("Cannot cast input to OfferWorkflowInput")
			return
		}

		oA := w.aP.GetActivity(activities.OriginateActivityName)
		if oA == nil {
			err = fmt.Errorf("activity %s not found", shared.ActionOriginate)
			return
		}
		ctx = workflow.WithActivityOptions(ctx,
			workflow.ActivityOptions{ScheduleToStartTimeout: time.Second, StartToCloseTimeout: wi.Timeout})

		ai := activities.OriginateActivityInput{}
		if ok := shared.ConvertInput(i, &ai); !ok {
			err = fserrors.NewWorkflowInputError("Cannot cast input to OriginateActivityInput")
			return
		}
		ai.Background = false

		for {
			err = workflow.ExecuteActivity(ctx, oA.Handler(), ai).Get(ctx, o)
			if err != nil {
				logger.Error("Failed to execute OriginateActivity", zap.Error(err))
				continue
			}

			res := o.Metadata[shared.FieldOutput]
			if res == nil {
				continue
			}

			if resStr, ok := res.(string); ok {
				if retryIgnored[resStr] {
					o.Success = true
					o.Metadata[shared.FieldUniqueId] = resStr
					break
				}
				if _, err := uuid.Parse(resStr); err == nil {
					o.Success = true
					o.Metadata[shared.FieldUniqueId] = resStr
					break
				}
			} else {
				break
			}
		}

		return
	}
}

func (w *OfferWorkflow) Name() string {
	return "workflows.OfferWorkflow"
}

func (w *OfferWorkflow) QueryResult(r shared.WorkflowQueryResult, e error) {
	if r != nil {
		for k, v := range r {
			r[k] = v
		}
	}

	if e != nil {
		w.e = e
	}
}

func (w *OfferWorkflow) SocketProvider() freeswitch.SocketProvider {
	return w.sP
}

func NewOfferWorkflow(sP freeswitch.SocketProvider, aP session.ActivityProvider) *OfferWorkflow {
	return &OfferWorkflow{sP: sP, aP: aP}
}

var _ shared.FreeswitchWorkflow = (*OfferWorkflow)(nil)
