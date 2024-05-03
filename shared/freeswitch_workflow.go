package shared

import "go.uber.org/cadence/workflow"

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
	CreateActivityProcessor(s string) (FreeswitchActivityProcessor, error)
}
