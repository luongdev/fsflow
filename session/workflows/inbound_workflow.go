package workflows

import (
	"github.com/luongdev/fsflow/errors"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/session/activities"
	"github.com/luongdev/fsflow/session/processors"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/workflow"
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
	sP freeswitch.SocketProvider
	aP session.ActivityProvider

	r shared.WorkflowQueryResult
	e error
}

func (w *InboundWorkflow) QueryResult(r shared.WorkflowQueryResult, e error) {
	if r != nil {
		for k, v := range r {
			r[k] = v
		}
	}

	if e != nil {
		w.e = e
	}
}

func (w *InboundWorkflow) SocketProvider() freeswitch.SocketProvider {
	return w.sP
}

func (w *InboundWorkflow) Name() string {
	return "workflows.InboundWorkflow"
}

func NewInboundWorkflow(sP freeswitch.SocketProvider, aP session.ActivityProvider) *InboundWorkflow {
	return &InboundWorkflow{sP: sP, aP: aP}
}

func (w *InboundWorkflow) Handler() shared.WorkflowFunc {
	return func(ctx workflow.Context, i shared.WorkflowInput) (*shared.WorkflowOutput, error) {
		logger := workflow.GetLogger(ctx)

		r := shared.WorkflowQueryResult{}
		var e error
		w.QueryResult(r, e)

		err := workflow.SetQueryHandler(ctx, string(shared.QuerySession), shared.NewQueryHandler(r, e))
		if err != nil {
			logger.Error("Failed to set query handler", zap.Error(err))
		}

		output := shared.NewWorkflowOutput(i.GetSessionId())
		if err := i.Validate(); err != nil {
			logger.Error("Invalid input", zap.Any("input", i), zap.Error(err))
			return output, err
		}

		input := InboundWorkflowInput{}
		ok := shared.ConvertInput(i, &input)

		if !ok {
			logger.Error("Failed to cast input to InboundWorkflowInput")
			return output, errors.NewWorkflowInputError("Cannot cast input to InboundWorkflowInput")
		}

		ctx = workflow.WithActivityOptions(ctx,
			workflow.ActivityOptions{ScheduleToStartTimeout: time.Second, StartToCloseTimeout: input.Timeout})

		si := w.aP.GetActivity("activities.SessionInitActivity")
		f := workflow.ExecuteActivity(ctx, si.Handler(), activities.SessionInitActivityInput{
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

		processor := processors.NewFreeswitchActivityProcessor(w, w.aP)
		output, err = processor.Process(ctx, output.Metadata)
		if err != nil {
			logger.Error("Failed to process metadata", zap.Any("metadata", output.Metadata), zap.Error(err))
		}

		r[shared.FieldAction] = output.Metadata.GetAction()
		r[shared.FieldInput] = output.Metadata.GetInput()

		m := shared.Metadata{}
		signalChan := workflow.GetSignalChannel(ctx, InboundSignal)
		for {
			s := workflow.NewSelector(ctx)
			s.AddReceive(signalChan, func(ch workflow.Channel, ok bool) {
				if ok {
					ch.Receive(ctx, &m)
				}
			})

			s.Select(ctx)

			r[shared.FieldAction] = m.GetAction()
			r[shared.FieldInput] = m.GetInput()

			if m.GetAction() == shared.ActionUnknown {
				//output.Metadata[shared.FieldAction] = shared.ActionHangup
				//output.Metadata[shared.FieldInput] = activities.HangupActivityInput{
				//	SessionId:    i.GetSessionId(),
				//	HangupReason: "InboundSignalUnknown",
				//	HangupCause:  "NORMAL_CLEARING",
				//}
				logger.Warn("Unknown signal. Waiting for other signals ...", zap.Any("metadata", m))
			} else {
				//ha := activities.NewHangupActivity(w.sP)
				//err = workflow.ExecuteActivity(ctx, ha.Handler(), activities.HangupActivityInput{
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
				//return output, err
				w.e = err
			} else {
				r[shared.FieldAction] = output.Metadata.GetAction()
				r[shared.FieldInput] = output.Metadata.GetInput()
			}
		}
	}
}

var _ shared.FreeswitchWorkflow = (*InboundWorkflow)(nil)
