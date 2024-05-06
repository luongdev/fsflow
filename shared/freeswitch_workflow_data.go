package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/cadence/workflow"
)

type Action string

const (
	ActionBridge    Action = "bridge"
	ActionAnswer    Action = "answer"
	ActionHangup    Action = "hangup"
	ActionTransfer  Action = "transfer"
	ActionOriginate Action = "originate"
	ActionUnknown   Action = "unknown"
)

type Field string

const (
	FieldAction    Field = "action"
	FieldMessage   Field = "message"
	FieldSessionId Field = "sessionId"
	FieldInput     Field = "input"
)

var actions = map[string]Action{
	string(ActionBridge):    ActionBridge,
	string(ActionAnswer):    ActionAnswer,
	string(ActionHangup):    ActionHangup,
	string(ActionTransfer):  ActionTransfer,
	string(ActionOriginate): ActionOriginate,
}

type Metadata map[Field]interface{}

func (m *Metadata) GetAction() Action {
	if v, ok := (*m)[FieldAction]; ok {
		if aStr, ok := v.(string); ok {
			if a, ok := actions[aStr]; ok {
				return a
			}
		}
	}

	return ActionUnknown
}

func (m *Metadata) GetSessionId() string {
	if v, ok := (*m)[FieldSessionId]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}

	return ""
}

func (m *Metadata) GetInput() WorkflowInput {
	if v, ok := (*m)[FieldInput]; ok {
		if i, ok := v.(map[string]interface{}); ok {
			input := WorkflowInput{}
			if ok := Convert(i, &input); ok {
				return input
			}
		}
	}

	return WorkflowInput{}
}

type WorkflowInput map[Field]interface{}

func (wi WorkflowInput) GetSessionId() string {
	i, ok := wi[FieldSessionId]
	if !ok {
		m, ok := wi["WorkflowInput"].(map[string]interface{})
		if !ok {
			return ""
		}
		i = m[string(FieldSessionId)]
	}

	return fmt.Sprintf("%v", i)
}

func (wi WorkflowInput) Validate() error {
	if wi.GetSessionId() == "" {
		return NewWorkflowInputError("sessionId is required")
	}
	return nil
}

type ActivityFunc func(ctx context.Context, i WorkflowInput) (*WorkflowOutput, error)

type WorkflowFunc func(ctx workflow.Context, i WorkflowInput) (*WorkflowOutput, error)

type WorkflowOutput struct {
	Success   bool     `json:"success"`
	SessionId string   `json:"sessionId"`
	Metadata  Metadata `json:"metadata"`
}

func NewWorkflowOutput(sessionId string) *WorkflowOutput {
	return &WorkflowOutput{
		Success:   false,
		SessionId: sessionId,
		Metadata:  Metadata{FieldSessionId: sessionId},
	}
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

func ConvertInput(in WorkflowInput, out interface{}) bool {
	if in["WorkflowInput"] == nil {
		in["WorkflowInput"] = WorkflowInput{FieldSessionId: in.GetSessionId()}
	}

	jsonData, err := json.Marshal(in)
	if err != nil {
		return false
	}

	err = json.Unmarshal(jsonData, out)
	if err != nil {
		return false
	}

	return true
}
