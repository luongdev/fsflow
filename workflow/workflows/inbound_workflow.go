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
	Initializer string        `json:"initializer"`
	Timeout     time.Duration `json:"timeout"`
	shared.WorkflowInput
}

const InboundSignal = "inbound"

type InboundWorkflow struct {
	p freeswitch.SocketProvider
}

func (w *InboundWorkflow) Name() string {
	return "workflows.InboundWorkflow"
}

func NewInboundWorkflow(p freeswitch.SocketProvider) *InboundWorkflow {
	return &InboundWorkflow{p: p}
}

func (w *InboundWorkflow) Handler() shared.WorkflowFunc {
	return func(ctx libworkflow.Context, i shared.WorkflowInput) (*shared.WorkflowOutput, error) {
		logger := libworkflow.GetLogger(ctx)
		output := shared.NewWorkflowOutput(i.GetSessionId())

		if err := i.Validate(); err != nil {
			logger.Error("Invalid input", zap.Any("input", i), zap.Error(err))
			return output, err
		}

		input := InboundWorkflowInput{}
		ok := shared.ConvertInput(i, &input)

		if !ok {
			logger.Error("Failed to cast input to InboundWorkflowInput")
			return output, shared.NewWorkflowInputError("Cannot cast input to InboundWorkflowInput")
		}

		ctx = libworkflow.WithActivityOptions(ctx, libworkflow.ActivityOptions{
			ScheduleToStartTimeout: time.Second,
			StartToCloseTimeout:    time.Hour,
		})

		si := activities.NewSessionInitActivity(w.p)
		f := libworkflow.ExecuteActivity(ctx, si.Handler(), activities.SessionInitActivityInput{
			ANI:         input.ANI,
			DNIS:        input.DNIS,
			Domain:      input.Domain,
			Initializer: input.Initializer,
			Timeout:     input.Timeout,
			SessionId:   i.GetSessionId(),
		})

		if err := f.Get(ctx, output); err != nil || !output.Success {
			logger.Error("Failed to execute SessionInitActivity", zap.Any("output", output), zap.Error(err))
			return output, err
		}

		processor := processors.NewFreeswitchActivityProcessor(w.p)
		output, err := processor.Process(ctx, output.Metadata)
		if err != nil {
			logger.Error("Failed to process metadata", zap.Any("metadata", output.Metadata), zap.Error(err))
		}

		m := shared.Metadata{}
		signalChan := libworkflow.GetSignalChannel(ctx, InboundSignal)
		for {
			s := libworkflow.NewSelector(ctx)
			s.AddReceive(signalChan, func(ch libworkflow.Channel, ok bool) {
				if ok {
					ch.Receive(ctx, &m)
				}
			})

			s.Select(ctx)

			if m.GetAction() == shared.ActionUnknown {
				output.Metadata[shared.FieldAction] = shared.ActionHangup
				output.Metadata[shared.FieldInput] = activities.HangupActivityInput{
					SessionId:    i.GetSessionId(),
					HangupReason: "InboundSignalUnknown",
					HangupCause:  "NORMAL_CLEARING",
				}
			} else {
				//ha := activities.NewHangupActivity(w.p)
				//err = libworkflow.ExecuteActivity(ctx, ha.Handler(), activities.HangupActivityInput{
				//	SessionId:    i.GetSessionId(),
				//	HangupReason: "InboundSignalUnknown",
				//	HangupCause:  "NORMAL_CLEARING",
				//}).Get(ctx, output)
				//
				//if err != nil || !output.Success {
				//	logger.Error("Failed to execute HangupActivity", zap.Any("output", output), zap.Error(err))
				//	return output, err
				//}
			}

			output, err := processor.Process(ctx, m)
			if err != nil || !output.Success {
				logger.Error("Failed to process metadata", zap.Any("metadata", output.Metadata), zap.Error(err))
				return output, err
			}
		}
	}
}

var _ shared.FreeswitchWorkflow = (*InboundWorkflow)(nil)
