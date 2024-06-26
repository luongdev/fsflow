package shared

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/luongdev/fsflow/errors"
	"go.uber.org/cadence/workflow"
	"time"
)

type Action string

const (
	ActionAnswer    Action = "answer"
	ActionBridge    Action = "bridge"
	ActionCallback  Action = "callback"
	ActionEvent     Action = "event"
	ActionHangup    Action = "hangup"
	ActionTransfer  Action = "transfer"
	ActionOriginate Action = "originate"
	ActionSet       Action = "set"
	ActionOffer     Action = "offer"
	ActionUnknown   Action = "unknown"
)

type Field string

const (
	FieldAction      Field = "action"
	FieldMessage     Field = "message"
	FieldSessionId   Field = "sessionId"
	FieldDomain      Field = "domain"
	FieldTimeout     Field = "timeout"
	FieldInput       Field = "input"
	FieldOutput      Field = "output"
	FieldUniqueId    Field = "uniqueId"
	FieldCallback    Field = "callback"
	FieldCurrentBLeg Field = "current_bleg"
)

var actions = map[string]Action{
	string(ActionAnswer):    ActionAnswer,
	string(ActionBridge):    ActionBridge,
	string(ActionCallback):  ActionCallback,
	string(ActionEvent):     ActionEvent,
	string(ActionHangup):    ActionHangup,
	string(ActionTransfer):  ActionTransfer,
	string(ActionOffer):     ActionOffer,
	string(ActionOriginate): ActionOriginate,
	string(ActionSet):       ActionSet,
}

type Query string

const (
	QuerySession Query = "session"
)

type Metadata map[Field]interface{}

func (m *Metadata) GetAction() Action {
	if v, ok := (*m)[FieldAction]; ok {
		if aStr, ok := v.(string); ok {
			if a, ok := actions[aStr]; ok {
				return a
			}
		}

		if aAct, ok := v.(Action); ok {
			return aAct
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

		if i, ok := v.(WorkflowInput); ok {
			return i
		}
	}

	return WorkflowInput{}
}

type WorkflowQueryResult map[Field]interface{}

type WorkflowQueryHandler func() (WorkflowQueryResult, error)

func NewQueryHandler(r WorkflowQueryResult, e error) WorkflowQueryHandler {
	return func() (WorkflowQueryResult, error) {
		return r, e
	}
}

type WorkflowCallback struct {
	URL     string                 `json:"url"`
	Method  string                 `json:"method"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
}

func (wc *WorkflowCallback) Validate() error {
	if wc.URL == "" {
		return errors.NewWorkflowInputError("url is required")
	}
	return nil
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

func (wi WorkflowInput) GetTimeout() time.Duration {
	t, ok := wi[FieldTimeout]
	if !ok {
		m, ok := wi["WorkflowInput"].(map[string]interface{})
		if !ok {
			return -1
		}
		t = m[string(FieldSessionId)]
	}

	if d, ok := t.(time.Duration); ok {
		return d
	}
	if f, ok := t.(float64); ok {
		return time.Duration(f)
	}

	return -1
}

func (wi WorkflowInput) GetCallback() *WorkflowCallback {
	cb, ok := wi[FieldCallback]
	if !ok {
		m, ok := wi["WorkflowInput"].(map[string]interface{})
		if !ok {
			return nil
		}
		if cb, ok = m[string(FieldCallback)]; !ok {
			return nil
		}
	}

	wc := &WorkflowCallback{}
	if ok := Convert(cb, wc); !ok {
		return nil
	}

	return wc
}

func (wi WorkflowInput) Validate() error {
	if wi.GetSessionId() == "" {
		return errors.NewWorkflowInputError("sessionId is required")
	}
	return nil
}

type WorkflowSignal struct {
	Action Action        `json:"action"`
	Input  WorkflowInput `json:"input"`
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
	if in == nil {
		return false
	}

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
