package processors

import (
	"fmt"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	libworkflow "go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type FreeswitchActivityProcessorImpl struct {
	FsClient *freeswitch.SocketClient
}

func NewFreeswitchActivityProcessor(client *freeswitch.SocketClient) *FreeswitchActivityProcessorImpl {
	return &FreeswitchActivityProcessorImpl{FsClient: client}
}

func (p *FreeswitchActivityProcessorImpl) Process(ctx libworkflow.Context, metadata shared.Metadata) (*shared.WorkflowOutput, error) {
	logger := libworkflow.GetLogger(ctx)
	o := &shared.WorkflowOutput{Success: false, Metadata: make(shared.Metadata)}
	if metadata == nil || metadata[shared.FieldAction] == nil {
		return o, shared.NewWorkflowInputError("metadata is nil")
	}

	sessionId := ctx.Value(shared.FieldSessionId).(string)
	logger.Info("Processing freeswitch activity", zap.String("sessionId", sessionId))

	factory := NewFreeswitchProcessorFactory(p.FsClient)

	processor, err := factory.CreateActivityProcessor(metadata[shared.FieldAction].(string))
	if err != nil {
		logger.Error("Failed to create activity processor", zap.Error(err))
		return o, err
	}

	o, err = processor.Process(ctx, metadata)
	if err != nil {
		return o, err
	}

	if o.Success && o.Metadata[shared.FieldAction] == nil {
		return o, nil
	}

	return p.Process(ctx, o.Metadata)
}

func (p *FreeswitchActivityProcessorImpl) GetInput(metadata shared.Metadata, i interface{}) error {
	if metadata == nil || metadata[shared.FieldAction] == nil {
		return fmt.Errorf("cannot found action")
	}

	ok := shared.Convert(metadata[shared.FieldInput], &i)
	if !ok {
		return fmt.Errorf("cannot cast input for action: %v", metadata[shared.FieldAction])
	}

	return nil
}

var _ shared.FreeswitchActivityProcessor = (*FreeswitchActivityProcessorImpl)(nil)
