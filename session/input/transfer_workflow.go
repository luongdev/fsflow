package input

import (
	"github.com/luongdev/fsflow/shared"
	"time"
)

type TransferWorkflowInput struct {
	shared.WorkflowInput

	Timeout   time.Duration          `json:"timeout"`
	Gateway   string                 `json:"gateway"`
	Profile   string                 `json:"profile"`
	ANI       string                 `json:"ani"`
	DNIS      string                 `json:"dnis"`
	OrigFrom  string                 `json:"origFrom"`
	OrigTo    string                 `json:"origTo"`
	Variables map[string]interface{} `json:"variables"`
	Callback  *WorkflowCallback      `json:"callback"`
}
