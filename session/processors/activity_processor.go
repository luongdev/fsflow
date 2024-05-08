package processors

import (
	"fmt"
	"github.com/luongdev/fsflow/errors"
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type FreeswitchActivityProcessorImpl struct {
	workflow shared.FreeswitchWorkflow

	aP session.ActivityProvider
}

func NewFreeswitchActivityProcessor(w shared.FreeswitchWorkflow, aP session.ActivityProvider) *FreeswitchActivityProcessorImpl {
	return &FreeswitchActivityProcessorImpl{workflow: w, aP: aP}
}

func (p *FreeswitchActivityProcessorImpl) Process(ctx workflow.Context, metadata shared.Metadata) (*shared.WorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	o := shared.NewWorkflowOutput(metadata.GetSessionId())
	var e error
	if metadata == nil || metadata.GetAction() == shared.ActionUnknown {
		e = errors.NewWorkflowInputError("metadata is nil")
		return o, e
	}

	if metadata.GetAction() == shared.ActionSet {
		r := shared.WorkflowQueryResult{}
		p.workflow.QueryResult(r, e)

		return o, e
	}

	factory := NewFreeswitchProcessorFactory(p.workflow, p.aP)
	processor, e := factory.CreateActivityProcessor(metadata.GetAction())
	if e != nil {
		logger.Error("Failed to create activity processor", zap.Error(e))
		return o, e
	}

	o, e = processor.Process(ctx, metadata)

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
