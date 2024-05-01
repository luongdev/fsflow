package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/cadence/workflow"
	"strconv"
)

type Action string

const (
	Bridge    Action = "bridge"
	Answer    Action = "answer"
	Hangup    Action = "hangup"
	Transfer  Action = "transfer"
	Originate Action = "originate"
)

type Field string

const (
	Destination Field = "destination"
	Timeout     Field = "timeout"
	Message     Field = "message"
	Uid         Field = "uid"
)

type Metadata map[Field]string

type ActivityFunc func(ctx context.Context, input interface{}) (WorkflowOutput, error)

type WorkflowFunc func(ctx workflow.Context, input interface{}) (WorkflowOutput, error)

func (m *Metadata) Set(key Field, value interface{}) {
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case bool:
		strValue = strconv.FormatBool(v)
	default:
		strValue = fmt.Sprintf("%v", v)
	}
	(*m)[key] = strValue
}

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
