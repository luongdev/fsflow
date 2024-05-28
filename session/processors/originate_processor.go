package processors

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/session/activities"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
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

	i := activities.OriginateActivityInput{}
	err := p.GetInput(metadata, &i)
	if err != nil {
		logger.Error("Failed to get input", zap.Error(err))
		return output, err
	}

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout:    i.Timeout,
		ScheduleToStartTimeout: 1,
	})

	oA := p.aP.GetActivity(activities.OriginateActivityName)
	err = workflow.ExecuteActivity(ctx, oA.Handler(), i).Get(ctx, &output)

	if err != nil {
		logger.Error("Failed to execute originate activity", zap.Error(err))
		return output, err
	}

	if output.Success {
		if !i.Background {
			uid, ok := output.Metadata[shared.FieldUniqueId].(string)
			if ok && uid != "" && i.Extension != "" && i.GetSessionId() != "" {
				output.Metadata[shared.FieldAction] = shared.ActionBridge
				bInput := activities.BridgeActivityInput{
					Originator:    i.GetSessionId(),
					Originatee:    uid,
					WorkflowInput: i.WorkflowInput,
				}
				output.Metadata[shared.FieldInput] = bInput

				if i.Direction == shared.Outbound {
					bInput.Originatee = i.GetSessionId()
					bInput.Originator = uid
				}
			}

			//go func() {
			//	err := p.sendCallback(i.Callback, metadata)
			//	if err != nil {
			//		logger.Error("Failed to send callback", zap.Error(err))
			//	}
			//}()
		}
	}

	return output, nil
}

func (p *OriginateProcessor) sendCallback(url string, i interface{}) error {
	bInput, err := json.Marshal(&i)
	if err != nil {
		return err
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, url, bytes.NewBuffer(bInput))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)

	defer func(res *http.Response) {
		if res != nil && res.Body != nil {
			err := res.Body.Close()
			if err != nil {
				log.Fatalf("failed to close response body %v", err)
			}
		}
	}(res)

	if err != nil {
		return err
	}

	if res != nil && res.StatusCode != http.StatusOK {
		return err
	}

	return nil
}

var _ shared.FreeswitchActivityProcessor = (*OriginateProcessor)(nil)
