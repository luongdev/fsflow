package shared

import (
	"go.uber.org/cadence/workflow"
)

type QueryResultWorkflow interface {
	FreeswitchWorkflow
	QueryResult(result WorkflowQueryResult, err error)
}

type FreeswitchWorkflow interface {
	Handler() WorkflowFunc
	Name() string
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
