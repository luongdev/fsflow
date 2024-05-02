package workflow

//
//import (
//	"github.com/luongdev/fsflow/freeswitch"
//	"github.com/luongdev/fsflow/shared"
//	"github.com/luongdev/fsflow/workflow/activities"
//	libworkflow "go.uber.org/cadence/workflow"
//)
//
//type Processor interface {
//	Process(ctx libworkflow.Context, metadata shared.Metadata) (shared.WorkflowOutput, error)
//}
//
//var _ Processor = (*FreeswitchActivityProcessor)(nil)
//
//type FreeswitchActivityProcessor[T interface{}] struct {
//	FsClient *freeswitch.SocketClient
//}
//
//func NewFreeswitchActivityProcessor[T interface{}](client *freeswitch.SocketClient) *FreeswitchActivityProcessor[T] {
//	return &FreeswitchActivityProcessor[T]{FsClient: client}
//}
//
//func (f *FreeswitchActivityProcessor[T]) GetInput(metadata shared.Metadata) (interface{}, error) {
//	if metadata == nil || metadata[shared.Action] == nil {
//		return nil, shared.NewWorkflowInputError("metadata is nil")
//	}
//
//	var i T
//	ok := shared.Convert(metadata[shared.Input], &i)
//	if !ok {
//		return nil, shared.NewWorkflowInputError("cannot cast input")
//	}
//
//	return i, nil
//}
//
//func (f *FreeswitchActivityProcessor[T]) Process(ctx libworkflow.Context, metadata shared.Metadata) (shared.WorkflowOutput, error) {
//	o := shared.WorkflowOutput{Success: false, Metadata: make(shared.Metadata)}
//	if metadata == nil || metadata[shared.Action] == nil {
//		return o, shared.NewWorkflowInputError("metadata is nil")
//	}
//
//	defer func() {
//		if r := recover(); r != nil {
//		}
//	}()
//
//	var p Processor
//
//	switch metadata[shared.Action].(shared.FsAction) {
//	case shared.Bridge:
//		p = activities.NewBridgeProcessor(f.FsClient)
//		break
//	default:
//		return o, shared.NewWorkflowInputError("unsupported action")
//	}
//
//	o, err := p.Process(ctx, metadata)
//	if err != nil {
//		return o, err
//	}
//
//	return f.Process(ctx, o.Metadata)
//}
