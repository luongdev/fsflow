package processors

import (
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
)

type FreeswitchProcessorFactoryImpl struct {
	fsClient *freeswitch.SocketClient
}

func NewFreeswitchProcessorFactory(fsClient *freeswitch.SocketClient) *FreeswitchProcessorFactoryImpl {
	return &FreeswitchProcessorFactoryImpl{fsClient: fsClient}
}

func (f *FreeswitchProcessorFactoryImpl) CreateActivityProcessor(s shared.Action) (shared.FreeswitchActivityProcessor, error) {
	switch s {
	case shared.ActionOriginate:
		return NewOriginateProcessor(f.fsClient), nil
	case shared.ActionBridge:
		return NewBridgeProcessor(f.fsClient), nil
	case shared.ActionHangup:
		return NewHangupProcessor(f.fsClient), nil

	default:
		return nil, shared.NewWorkflowInputError("unsupported action")
	}
}

var _ shared.FreeswitchProcessorFactory = (*FreeswitchProcessorFactoryImpl)(nil)
