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

func (f *FreeswitchProcessorFactoryImpl) CreateActivityProcessor(s string) (shared.FreeswitchActivityProcessor, error) {
	switch s {
	case string(shared.Originate):
		return NewOriginateProcessor(f.fsClient), nil
	case string(shared.Bridge):
		return NewBridgeProcessor(f.fsClient), nil
	case string(shared.Hangup):
		return NewHangupProcessor(f.fsClient), nil

	default:
		return nil, shared.NewWorkflowInputError("unsupported action")
	}
}

var _ shared.FreeswitchProcessorFactory = (*FreeswitchProcessorFactoryImpl)(nil)
