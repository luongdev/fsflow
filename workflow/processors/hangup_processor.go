package processors

import (
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"github.com/luongdev/fsflow/workflow/activities"
	libworkflow "go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type HangupProcessor struct {
	*FreeswitchActivityProcessorImpl
}

func NewHangupProcessor(client freeswitch.SocketProvider) *HangupProcessor {
	return &HangupProcessor{FreeswitchActivityProcessorImpl: NewFreeswitchActivityProcessor(client)}
}

func (p *HangupProcessor) Process(ctx libworkflow.Context, metadata shared.Metadata) (*shared.WorkflowOutput, error) {
	logger := libworkflow.GetLogger(ctx)
	output := shared.NewWorkflowOutput(metadata.GetSessionId())

	i := activities.HangupActivityInput{}
	err := p.GetInput(metadata, &i)
	if err != nil {
		logger.Error("Failed to get input", zap.Error(err))
		return output, err
	}

	hangupActivity := activities.NewHangupActivity(p.SocketProvider)
	err = libworkflow.ExecuteActivity(ctx, hangupActivity.Handler(), i).Get(ctx, &output)

	return output, err
}

var _ shared.FreeswitchActivityProcessor = (*BridgeProcessor)(nil)
