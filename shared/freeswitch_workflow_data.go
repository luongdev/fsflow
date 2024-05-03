package shared

import (
	"context"
	"encoding/json"
	"go.uber.org/cadence/workflow"
)

type FsAction string

const (
	Bridge    FsAction = "bridge"
	Answer    FsAction = "answer"
	Hangup    FsAction = "hangup"
	Transfer  FsAction = "transfer"
	Originate FsAction = "originate"
)

type Field string

const (
	Action  Field = "action"
	Message Field = "message"
	Uid     Field = "uid"
	Input   Field = "input"
)

type Metadata map[Field]interface{}

type ActivityFunc func(ctx context.Context, i interface{}) (WorkflowOutput, error)

type WorkflowFunc func(ctx workflow.Context, i interface{}) (WorkflowOutput, error)

type WorkflowOutput struct {
	Success  bool     `json:"success"`
	Metadata Metadata `json:"metadata"`
}

func Convert(m interface{}, target interface{}) bool {
	jsonData, err := json.Marshal(m)
	if err != nil {
		return false
	}

	err = json.Unmarshal(jsonData, target)
	if err != nil {
		return false
	}

	return true
}
