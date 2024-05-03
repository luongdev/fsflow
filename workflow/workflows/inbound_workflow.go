package workflows

import (
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"github.com/luongdev/fsflow/workflow/activities"
	"github.com/luongdev/fsflow/workflow/processors"
	libworkflow "go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"time"
)

type InboundWorkflowInput struct {
	ANI         string        `json:"ani"`
	DNIS        string        `json:"dnis"`
	Domain      string        `json:"domain"`
	SessionId   string        `json:"sessionId"`
	Initializer string        `json:"initializer"`
	Timeout     time.Duration `json:"timeout"`
}

const InboundSignal = "inbound"

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
	return func(ctx libworkflow.Context, i interface{}) (shared.WorkflowOutput, error) {
		logger := libworkflow.GetLogger(ctx)
		output := shared.WorkflowOutput{Success: false, Metadata: make(shared.Metadata)}
		input := InboundWorkflowInput{}
		ok := shared.Convert(i, &input)

		if !ok {
			logger.Error("Failed to cast input to InboundWorkflowInput")
			return output, shared.NewWorkflowInputError("Cannot cast input to InboundWorkflowInput")
		}

		ctx = libworkflow.WithActivityOptions(
			ctx,
			libworkflow.ActivityOptions{ScheduleToStartTimeout: time.Second, StartToCloseTimeout: input.Timeout},
		)
		ctx = libworkflow.WithValue(ctx, shared.Uid, input.SessionId)

		siActivity := activities.NewSessionInitActivity(w.fsClient)
		f := libworkflow.ExecuteActivity(ctx, siActivity.Handler(), activities.SessionInitActivityInput{
			ANI:         input.ANI,
			DNIS:        input.DNIS,
			Domain:      input.Domain,
			Initializer: input.Initializer,
			Timeout:     input.Timeout,
			SessionId:   input.SessionId,
		})

		if err := f.Get(ctx, &output); err != nil || !output.Success {
			logger.Error("Failed to execute SessionInitActivity", zap.Any("output", output), zap.Error(err))
			return output, err
		}

		processor := processors.NewFreeswitchActivityProcessor(w.fsClient)
		ctx = libworkflow.WithValue(ctx, shared.Uid, input.SessionId)
		output, err := processor.Process(ctx, output.Metadata)
		if err != nil || !output.Success {
			logger.Error("Failed to process metadata", zap.Any("metadata", output.Metadata), zap.Error(err))
			return output, err
		}

		//switch output.Metadata[shared.Action].(string) {
		//case string(shared.Bridge):
		//	break
		//case string(shared.Hangup):
		//	hupActivity := activities.NewHangupActivity(w.fsClient)
		//	hi := activities.HangupActivityInput{SessionId: input.SessionId}
		//	if output.Metadata[shared.HangupCause] != nil {
		//		hi.HangupCause = output.Metadata[shared.HangupCause].(string)
		//	}
		//	err := libworkflow.ExecuteActivity(ctx, hupActivity.Handler(), hi).Get(ctx, &output)
		//	if err != nil {
		//		logger.Error("Failed to execute HangupActivity", zap.Any("output", output), zap.Error(err))
		//		return output, err
		//	}
		//	break
		//case string(shared.Originate):
		//	if output.Metadata[shared.Destination] == nil {
		//		logger.Error("Missing required metadata", zap.Any("output", output))
		//		return output, shared.RequireField("destination")
		//	}
		//
		//	if output.Metadata[shared.Gateway] == nil {
		//		logger.Error("Missing required metadata", zap.Any("output", output))
		//		return output, shared.RequireField("gateway")
		//	}
		//
		//	oi := activities.OriginateActivityInput{
		//		Timeout:     input.Timeout,
		//		Destination: output.Metadata[shared.Destination].(string),
		//		Gateway:     output.Metadata[shared.Gateway].(string),
		//		AllowReject: true,
		//		AutoAnswer:  false,
		//		Direction:   freeswitch.Inbound,
		//	}
		//	if output.Metadata[shared.Profile] != nil {
		//		oi.Profile = output.Metadata[shared.Profile].(string)
		//	}
		//
		//	origActivity := activities.NewOriginateActivity(w.fsClient)
		//	err := libworkflow.ExecuteActivity(ctx, origActivity.Handler(), oi).Get(ctx, &output)
		//	if err != nil || !output.Success {
		//		logger.Error("Failed to execute OriginateActivity", zap.Any("output", output), zap.Error(err))
		//		return output, err
		//	}
		//
		//	if output.Metadata[shared.Uid] == nil {
		//		logger.Error("Missing required metadata", zap.Any("output", output))
		//		return output, shared.RequireField("uid")
		//	}
		//
		//	brActivity := activities.NewBridgeActivity(w.fsClient)
		//	bi := activities.BridgeActivityInput{
		//		Originator: input.SessionId,
		//		Originatee: output.Metadata[shared.Uid].(string),
		//	}
		//
		//	output.Metadata[shared.Input] = bi
		//	//p := workflow.NewFreeswitchActivityProcessor[interface{}](w.fsClient)
		//	//o, err := p.Process(ctx, output.Metadata)
		//	//if err != nil || !o.Success {
		//	//	logger.Error("Failed to process metadata", zap.Any("metadata", output.Metadata), zap.Error(err))
		//	//	return output, err
		//	//}
		//
		//	err = libworkflow.ExecuteActivity(ctx, brActivity.Handler(), bi).Get(ctx, &output)
		//	if err != nil || !output.Success {
		//		logger.Error("Failed to execute BridgeActivity", zap.Any("output", output), zap.Error(err))
		//		return output, err
		//	}
		//
		//	break
		//default:
		//	break
		//}

		signalChan := libworkflow.GetSignalChannel(ctx, InboundSignal)
		for {
			s := libworkflow.NewSelector(ctx)
			s.AddReceive(signalChan, func(ch libworkflow.Channel, ok bool) {
				if ok {
					ch.Receive(ctx, output)
				}
			})

			s.Select(ctx)

			if output.Success {
				if output.Metadata[shared.Action] == shared.Hangup {
					result := shared.WorkflowOutput{}
					//_ = lib_workflow.ExecuteActivity(ctx, cmdActivities.HangupActivity, activities.HangupActivityInput{
					//	SessionId:   input.SessionId,
					//	HangupCause: "CALL_REJECTED",
					//}).Get(ctx, &result)

					return result, nil
				}
			}
		}
	}
}

var _ shared.FreeswitchWorkflow = (*InboundWorkflow)(nil)
