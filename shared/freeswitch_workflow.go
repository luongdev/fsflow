package shared

import (
	"github.com/luongdev/fsflow/freeswitch"
	"go.uber.org/cadence/workflow"
)

type FreeswitchWorkflow interface {
	Handler() WorkflowFunc
	Name() string

	QueryResult(r WorkflowQueryResult, e error)
	SocketProvider() freeswitch.SocketProvider
}

type FreeswitchActivity interface {
	Handler() ActivityFunc

	Name() string
}

type FreeswitchActivityProcessor interface {
	Process(ctx workflow.Context, metadata Metadata) (*WorkflowOutput, error)
}

type FreeswitchProcessorFactory interface {
	CreateActivityProcessor(s Action) (FreeswitchActivityProcessor, error)
}
