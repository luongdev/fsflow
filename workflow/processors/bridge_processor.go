package processors

import (
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"github.com/luongdev/fsflow/workflow/activities"
	libworkflow "go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type BridgeProcessor struct {
	*FreeswitchActivityProcessorImpl
}

func NewBridgeProcessor(client *freeswitch.SocketClient) *BridgeProcessor {
	return &BridgeProcessor{FreeswitchActivityProcessorImpl: NewFreeswitchActivityProcessor(client)}
}

func (p *BridgeProcessor) Process(ctx libworkflow.Context, metadata shared.Metadata) (shared.WorkflowOutput, error) {
	logger := libworkflow.GetLogger(ctx)
	output := shared.WorkflowOutput{Success: false, Metadata: make(shared.Metadata)}

	i := activities.BridgeActivityInput{}
	err := p.GetInput(metadata, &i)
	if err != nil {
		logger.Error("Failed to get input", zap.Error(err))
		return output, err
	}

	bridgeActivity := activities.NewBridgeActivity(p.FsClient)
	err = libworkflow.ExecuteActivity(ctx, bridgeActivity.Handler(), i).Get(ctx, &output)

	return output, err
}

var _ shared.FreeswitchActivityProcessor = (*BridgeProcessor)(nil)
