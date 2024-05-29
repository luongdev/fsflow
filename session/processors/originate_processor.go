package processors

import (
	"encoding/json"
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/session/activities"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type OriginateProcessor struct {
	*FreeswitchActivityProcessorImpl
}

func NewOriginateProcessor(w shared.FreeswitchWorkflow, aP session.ActivityProvider) *OriginateProcessor {
	return &OriginateProcessor{FreeswitchActivityProcessorImpl: NewFreeswitchActivityProcessor(w, aP)}
}

func (p *OriginateProcessor) Process(ctx workflow.Context, metadata shared.Metadata) (*shared.WorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	output := shared.NewWorkflowOutput(metadata.GetSessionId())

	oi := activities.OriginateActivityInput{}
	err := p.GetInput(metadata, &oi)
	if err != nil {
		logger.Error("Failed to get input", zap.Error(err))
		return output, err
	}

	if oi.Background {
		if cb := metadata.GetInput().GetCallback(); cb != nil {
			if oi.Variables == nil {
				oi.Variables = make(map[string]interface{})
			}

			oi.Variables["callback_url"] = cb.URL
			if cb.Method != "" {
				oi.Variables["callback_method"] = cb.Method
			}
			if cb.Headers != nil && len(cb.Headers) > 0 {
				if h, err := json.Marshal(cb.Headers); err != nil {
					oi.Variables["callback_headers"] = string(h)
				}
			}
			if cb.Body != nil && len(cb.Body) > 0 {
				if b, err := json.Marshal(cb.Body); err != nil {
					oi.Variables["callback_body"] = string(b)
				}
			}
		}
	}

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout:    oi.Timeout,
		ScheduleToStartTimeout: 1,
	})

	oA := p.aP.GetActivity(activities.OriginateActivityName)
	err = workflow.ExecuteActivity(ctx, oA.Handler(), oi).Get(ctx, &output)

	if err != nil {
		logger.Error("Failed to execute originate activity", zap.Error(err))
		return output, err
	}

	if output.Success {
		if !oi.Background {
			uid, ok := output.Metadata[shared.FieldUniqueId].(string)
			if ok && uid != "" && oi.Extension != "" && oi.GetSessionId() != "" {
				output.Metadata[shared.FieldAction] = shared.ActionBridge
				bInput := activities.BridgeActivityInput{
					Originator:    oi.GetSessionId(),
					Originatee:    uid,
					WorkflowInput: oi.WorkflowInput,
				}

				if bInput.WorkflowInput != nil {
					if cb := metadata.GetInput().GetCallback(); cb != nil {
						bInput.WorkflowInput[shared.FieldCallback] =
							&shared.WorkflowCallback{URL: cb.URL, Method: cb.Method, Headers: cb.Headers, Body: cb.Body}
					}
				}

				output.Metadata[shared.FieldInput] = bInput
				if oi.Direction == shared.Outbound {
					bInput.Originatee = oi.GetSessionId()
					bInput.Originator = uid
				}
			}
		}
	}

	return output, nil
}

var _ shared.FreeswitchActivityProcessor = (*OriginateProcessor)(nil)
