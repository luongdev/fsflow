package processors

import (
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
)

type FreeswitchProcessorFactoryImpl struct {
	socketProvider freeswitch.SocketProvider
}

func NewFreeswitchProcessorFactory(provider freeswitch.SocketProvider) *FreeswitchProcessorFactoryImpl {
	return &FreeswitchProcessorFactoryImpl{socketProvider: provider}
}

func (f *FreeswitchProcessorFactoryImpl) CreateActivityProcessor(s shared.Action) (shared.FreeswitchActivityProcessor, error) {
	switch s {
	case shared.ActionOriginate:
		return NewOriginateProcessor(f.socketProvider), nil
	case shared.ActionBridge:
		return NewBridgeProcessor(f.socketProvider), nil
	case shared.ActionHangup:
		return NewHangupProcessor(f.socketProvider), nil
	case shared.ActionEvent:
		return NewEventProcessor(f.socketProvider), nil

	default:
		return nil, shared.NewWorkflowInputError("unsupported action")
	}
}

var _ shared.FreeswitchProcessorFactory = (*FreeswitchProcessorFactoryImpl)(nil)
