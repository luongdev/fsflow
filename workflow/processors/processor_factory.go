package processors

import (
	error2 "github.com/luongdev/fsflow/errors"
	"github.com/luongdev/fsflow/shared"
)

type FreeswitchProcessorFactoryImpl struct {
	workflow shared.FreeswitchWorkflow
}

func NewFreeswitchProcessorFactory(w shared.FreeswitchWorkflow) *FreeswitchProcessorFactoryImpl {
	return &FreeswitchProcessorFactoryImpl{workflow: w}
}

func (f *FreeswitchProcessorFactoryImpl) CreateActivityProcessor(s shared.Action) (shared.FreeswitchActivityProcessor, error) {
	switch s {
	case shared.ActionOriginate:
		return NewOriginateProcessor(f.workflow), nil
	case shared.ActionBridge:
		return NewBridgeProcessor(f.workflow), nil
	case shared.ActionHangup:
		return NewHangupProcessor(f.workflow), nil
	case shared.ActionEvent:
		return NewEventProcessor(f.workflow), nil

	default:
		return nil, error2.NewWorkflowInputError("unsupported action")
	}
}

var _ shared.FreeswitchProcessorFactory = (*FreeswitchProcessorFactoryImpl)(nil)
