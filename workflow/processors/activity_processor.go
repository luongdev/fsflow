package processors

import (
	"fmt"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type FreeswitchActivityProcessorImpl struct {
	FsClient *freeswitch.SocketClient
}

func NewFreeswitchActivityProcessor(client *freeswitch.SocketClient) *FreeswitchActivityProcessorImpl {
	return &FreeswitchActivityProcessorImpl{FsClient: client}
}

func (p *FreeswitchActivityProcessorImpl) Process(ctx workflow.Context, metadata shared.Metadata) (shared.WorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	o := shared.WorkflowOutput{Success: false, Metadata: make(shared.Metadata)}
	if metadata == nil || metadata[shared.Action] == nil {
		return o, shared.NewWorkflowInputError("metadata is nil")
	}

	sessionId := ctx.Value(shared.Uid).(string)
	logger.Info("Processing freeswitch activity", zap.String("sessionId", sessionId))

	var processor shared.FreeswitchActivityProcessor

	switch metadata[shared.Action].(string) {
	case string(shared.Originate):
		processor = NewOriginateProcessor(p.FsClient)
	case string(shared.Bridge):
		processor = NewBridgeProcessor(p.FsClient)
	}

	if processor == nil {
		return o, shared.NewWorkflowInputError("unsupported action")
	}

	o, err := processor.Process(ctx, metadata)
	if err != nil {
		return o, err
	}

	if o.Success && o.Metadata[shared.Action] == nil {
		return o, nil
	}

	return p.Process(ctx, o.Metadata)
}

func (p *FreeswitchActivityProcessorImpl) GetInput(metadata shared.Metadata, i interface{}) error {
	if metadata == nil || metadata[shared.Action] == nil {
		return fmt.Errorf("cannot found action")
	}

	ok := shared.Convert(metadata[shared.Input], &i)
	if !ok {
		return fmt.Errorf("cannot cast input for action: %v", metadata[shared.Action])
	}

	return nil
}

var _ shared.FreeswitchActivityProcessor = (*FreeswitchActivityProcessorImpl)(nil)

func errorRecover(ec chan error) {
	if r := recover(); r != nil {
		switch x := r.(type) {
		case error:
			ec <- x
		default:
			ec <- fmt.Errorf("%v", x)
		}
	}
}
