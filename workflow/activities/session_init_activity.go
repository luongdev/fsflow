package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type SessionInitActivityInput struct {
	ANI         string        `json:"ani"`
	DNIS        string        `json:"dnis"`
	Domain      string        `json:"domain"`
	Initializer string        `json:"initializer"`
	Timeout     time.Duration `json:"timeout"`
	SessionId   string        `json:"sessionId"`
}

type SessionInitActivity struct {
	fsClient *freeswitch.SocketClient
}

func NewSessionInitActivity(fsClient *freeswitch.SocketClient) *SessionInitActivity {
	return &SessionInitActivity{fsClient: fsClient}
}

func (s SessionInitActivity) Name() string {
	return "activities.SessionInitActivity"
}

func (s SessionInitActivity) Handler() shared.ActivityFunc {
	return func(ctx context.Context, i shared.WorkflowInput) (*shared.WorkflowOutput, error) {
		logger := activity.GetLogger(ctx)
		output := shared.NewWorkflowOutput(i.GetSessionId())

		if err := i.Validate(); err != nil {
			logger.Error("Invalid input", zap.Any("input", i), zap.Error(err))
			return output, err
		}

		input := SessionInitActivityInput{}
		ok := shared.ConvertInput(i, &input)

		if !ok {
			logger.Error("Failed to cast input to SessionInitActivityInput")
			return output, shared.NewWorkflowInputError("Cannot cast input to SessionInitActivityInput")
		}

		bInput, err := json.Marshal(&input)
		if err != nil {
			logger.Error("Failed to marshal input", zap.Error(err))
			return output, err
		}

		reqCtx, cancel := context.WithTimeout(ctx, input.Timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, input.Initializer, bytes.NewBuffer(bInput))
		if err != nil {
			logger.Error("Failed to create request to init session", zap.Error(err))
			return output, err
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(req)

		defer func(res *http.Response) {
			if res != nil && res.Body != nil {
				err := res.Body.Close()
				if err != nil {
					logger.Error("Failed to close response body", zap.Error(err))
				}
			}
		}(res)

		if err != nil {
			logger.Error("Failed to send request to initializer", zap.Error(err))
			return output, err
		}

		if res != nil && res.StatusCode != http.StatusOK {
			logger.Error("Failed to init session", zap.Any("status", res.StatusCode))
			return output, shared.NewWorkflowInputError("Failed to init session")
		}

		var o interface{}
		err = json.NewDecoder(res.Body).Decode(&o)
		if err != nil {
			logger.Error("Failed to decode response body", zap.Error(err))
			return output, err
		}

		if ok := shared.Convert(o, &output); !ok {
			logger.Error("Failed to cast response to WorkflowOutput")
			return output, shared.NewWorkflowInputError("Cannot cast response to WorkflowOutput")
		}

		return output, nil
	}
}

var _ shared.FreeswitchActivity = (*SessionInitActivity)(nil)
