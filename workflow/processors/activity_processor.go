package processors

import (
	"fmt"
	"github.com/luongdev/fsflow/errors"
	"github.com/luongdev/fsflow/shared"
	libworkflow "go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type FreeswitchActivityProcessorImpl struct {
	workflow shared.FreeswitchWorkflow
}

func NewFreeswitchActivityProcessor(w shared.FreeswitchWorkflow) *FreeswitchActivityProcessorImpl {
	return &FreeswitchActivityProcessorImpl{workflow: w}
}

func (p *FreeswitchActivityProcessorImpl) Process(ctx libworkflow.Context, metadata shared.Metadata) (*shared.WorkflowOutput, error) {
	logger := libworkflow.GetLogger(ctx)
	o := shared.NewWorkflowOutput(metadata.GetSessionId())
	if metadata == nil || metadata.GetAction() == shared.ActionUnknown {
		return o, errors.NewWorkflowInputError("metadata is nil")
	}

	factory := NewFreeswitchProcessorFactory(p.workflow)
	processor, err := factory.CreateActivityProcessor(metadata.GetAction())
	if err != nil {
		logger.Error("Failed to create activity processor", zap.Error(err))
		return o, err
	}

	o, err = processor.Process(ctx, metadata)

	if o != nil && o.Metadata.GetAction() != shared.ActionUnknown {
		return p.Process(ctx, o.Metadata)
	}

	return o, nil
}

func (p *FreeswitchActivityProcessorImpl) GetInput(metadata shared.Metadata, i interface{}) error {
	if metadata == nil || metadata.GetAction() == shared.ActionUnknown {
		return fmt.Errorf("cannot found action")
	}

	ok := shared.ConvertInput(metadata.GetInput(), &i)
	if !ok {
		return fmt.Errorf("cannot cast input for action: %v", metadata.GetAction())
	}

	return nil
}

var _ shared.FreeswitchActivityProcessor = (*FreeswitchActivityProcessorImpl)(nil)
