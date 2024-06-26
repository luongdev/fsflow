package input

import (
	"github.com/luongdev/fsflow/shared"
	"time"
)

type TransferWorkflowInput struct {
	shared.WorkflowInput

	Destination string        `json:"destination"`
	Timeout     time.Duration `json:"timeout"`
}
