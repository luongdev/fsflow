package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/luongdev/fsflow/errors"
	"github.com/luongdev/fsflow/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type Method string

const (
	MethodGet    Method = "GET"
	MethodPost   Method = "POST"
	MethodPut    Method = "PUT"
	MethodDelete Method = "DELETE"
)

func (m Method) HttpMethod() string {
	switch m {
	case MethodGet:
		return http.MethodGet
	case MethodPost:
		return http.MethodPost
	case MethodPut:
		return http.MethodPut
	case MethodDelete:
		return http.MethodDelete
	default:
		return http.MethodGet
	}
}

type CallbackActivityInput struct {
	shared.WorkflowInput

	Timeout time.Duration          `json:"timeout"`
	URL     string                 `json:"url"`
	Method  Method                 `json:"method"`
	Queries map[string]interface{} `json:"queries"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
}

type CallbackActivity struct {
}

const CallbackActivityName = "activities.CallbackActivity"

func (c *CallbackActivity) Name() string {
	return CallbackActivityName
}

func NewCallbackActivity() *CallbackActivity {
	return &CallbackActivity{}
}

func (c *CallbackActivity) Handler() shared.ActivityFunc {
	return func(ctx context.Context, i shared.WorkflowInput) (*shared.WorkflowOutput, error) {
		logger := activity.GetLogger(ctx)
		output := shared.NewWorkflowOutput(i.GetSessionId())

		if err := i.Validate(); err != nil {
			logger.Error("Invalid input", zap.Any("input", i), zap.Error(err))
			return output, err
		}

		input := CallbackActivityInput{}
		ok := shared.ConvertInput(i, &input)

		if !ok {
			logger.Error("Failed to cast input to CallbackActivityInput")
			return output, fmt.Errorf("cannot cast input to CallbackActivityInput")
		}

		bInput, err := json.Marshal(input.Body)
		if err != nil {
			logger.Error("Failed to marshal input", zap.Error(err))
			return output, err
		}

		reqCtx, cancel := context.WithTimeout(ctx, input.Timeout)
		defer cancel()

		q := ""
		for k, v := range input.Queries {
			q += fmt.Sprintf("&%s=%v", k, v)
		}
		if q != "" {
			input.URL += "?" + q
		}

		req, err := http.NewRequestWithContext(reqCtx, input.Method.HttpMethod(), input.URL, bytes.NewBuffer(bInput))
		if err != nil {
			logger.Error("Failed to create callback request", zap.Error(err))
			return output, err
		}

		req.Header.Set("Content-Type", "application/json")
		for k, v := range input.Headers {
			req.Header.Set(k, v)
		}

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
			return output, errors.NewWorkflowInputError("Failed to init session")
		}

		var o interface{}
		err = json.NewDecoder(res.Body).Decode(&o)
		if err != nil {
			logger.Error("Failed to decode response body", zap.Error(err))
			return output, err
		}
		output.Success = true
		output.Metadata[shared.FieldOutput] = o

		return output, nil
	}
}

var _ shared.FreeswitchActivity = (*CallbackActivity)(nil)
