package processors

import (
	"github.com/luongdev/fsflow/errors"
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/shared"
)

type FreeswitchProcessorFactoryImpl struct {
	workflow shared.FreeswitchWorkflow

	aP session.ActivityProvider
}

func NewFreeswitchProcessorFactory(w shared.FreeswitchWorkflow, aP session.ActivityProvider) *FreeswitchProcessorFactoryImpl {
	return &FreeswitchProcessorFactoryImpl{workflow: w, aP: aP}
}

func (f *FreeswitchProcessorFactoryImpl) CreateActivityProcessor(s shared.Action) (shared.FreeswitchActivityProcessor, error) {
	switch s {
	case shared.ActionOriginate:
		return NewOriginateProcessor(f.workflow, f.aP), nil
	case shared.ActionBridge:
		return NewBridgeProcessor(f.workflow, f.aP), nil
	case shared.ActionHangup:
		return NewHangupProcessor(f.workflow, f.aP), nil
	case shared.ActionEvent:
		return NewEventProcessor(f.workflow, f.aP), nil

	default:
		return nil, errors.NewWorkflowInputError("unsupported action")
	}
}

var _ shared.FreeswitchProcessorFactory = (*FreeswitchProcessorFactoryImpl)(nil)
