package input

import (
	"github.com/google/uuid"
	"github.com/luongdev/fsflow/shared"
	"time"
)

type OfferWorkflowInput struct {
	shared.WorkflowInput

	UId         uuid.UUID              `json:"uid"`
	Timeout     time.Duration          `json:"timeout"`
	Gateway     string                 `json:"gateway"`
	Profile     string                 `json:"profile"`
	ANI         string                 `json:"ani"`
	DNIS        string                 `json:"dnis"`
	OrigFrom    string                 `json:"origFrom"`
	OrigTo      string                 `json:"origTo"`
	AutoAnswer  bool                   `json:"autoAnswer"`
	AllowReject bool                   `json:"allowReject"`
	Direction   string                 `json:"direction"`
	Variables   map[string]interface{} `json:"variables"`
	Extension   string                 `json:"extension"`
	Callback    *WorkflowCallback      `json:"callback"`
}
