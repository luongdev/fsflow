package workflows

import (
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"github.com/luongdev/fsflow/workflow/activities"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"time"
)

type InboundWorkflowInput struct {
	PhoneNumber  string `json:"phoneNumber"`
	DialedNumber string `json:"dialedNumber"`
	SessionId    string `json:"sessionId"`
	Initializer  string `json:"initializer"`
}

const InitCompletedSignal = "init_completed"

type InboundWorkflow struct {
	fsClient *freeswitch.SocketClient
}

func (w *InboundWorkflow) Name() string {
	return "workflows.InboundWorkflow"
}

func NewInboundWorkflow(fsClient *freeswitch.SocketClient) *InboundWorkflow {
	return &InboundWorkflow{fsClient: fsClient}
}

func (w *InboundWorkflow) Handler() shared.WorkflowFunc {
	return func(ctx workflow.Context, i interface{}) (shared.WorkflowOutput, error) {
		logger := workflow.GetLogger(ctx)
		output := shared.WorkflowOutput{Success: false, Metadata: make(shared.Metadata)}
		input := InboundWorkflowInput{}
		ok := shared.Convert(i, &input)

		if !ok {
			logger.Error("Failed to cast input to InboundWorkflowInput")
			return output, shared.NewWorkflowInputError("Cannot cast input to InboundWorkflowInput")
		}

		ao := workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
		}

		ctx = workflow.WithActivityOptions(ctx, ao)

		hangupActivity := activities.NewHangupActivity(w.fsClient)
		future := workflow.ExecuteActivity(ctx, hangupActivity.Handler(), activities.HangupActivityInput{
			SessionId:   input.SessionId,
			HangupCause: "CALL_REJECTED",
		})

		err := future.Get(ctx, &output)
		if err != nil {
			logger.Error("Failed to execute hangup activity", zap.Error(err))
			return output, err
		}

		return output, nil
	}
}

var _ shared.FreeswitchWorkflow = (*InboundWorkflow)(nil)
