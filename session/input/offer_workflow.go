package input

import (
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"time"
)

type OfferWorkflowInput struct {
	shared.WorkflowInput

	Timeout      time.Duration          `json:"timeout"`
	DialedNumber string                 `json:"dialedNumber"`
	Destination  string                 `json:"destination"`
	Gateway      string                 `json:"gateway"`
	Profile      string                 `json:"profile"`
	ANI          string                 `json:"ani"`
	DNIS         string                 `json:"dnis"`
	AutoAnswer   bool                   `json:"autoAnswer"`
	AllowReject  bool                   `json:"allowReject"`
	Direction    freeswitch.Direction   `json:"direction"`
	Variables    map[string]interface{} `json:"variables"`
	Extension    string                 `json:"extension"`
	Callback     string                 `json:"callback"`
}
