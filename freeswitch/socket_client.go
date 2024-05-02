package freeswitch

import (
	"context"
	"fmt"
	"github.com/luongdev/fsflow/shared"
	"github.com/percipia/eslgo"
	"github.com/percipia/eslgo/command"
	"github.com/percipia/eslgo/command/call"
	"time"
)

var _ SocketClient = (*SocketClientImpl)(nil)

type SocketClientImpl struct {
	*eslgo.Conn
}

func NewSocketClient(conn *eslgo.Conn) *SocketClientImpl {
	return &SocketClientImpl{conn}
}

func (s *SocketClientImpl) Execute(ctx context.Context, cmd *Command) (string, error) {
	if cmd.Uid == "" {
		return "", fmt.Errorf("uuid is required")
	}

	raw, err := s.Conn.SendCommand(ctx, &call.Execute{
		UUID:    cmd.Uid,
		AppName: cmd.AppName,
		AppArgs: cmd.AppArgs,
	})

	if err != nil {
		return "", err
	}

	res, ok := NewResponse(raw).Get()
	if !ok {
		return res, fmt.Errorf("failed to execute command '%v': %v", cmd.AppName, res)
	}

	return res, nil
}

func (s *SocketClientImpl) Api(ctx context.Context, cmd *Command) (string, error) {
	raw, err := s.Conn.SendCommand(ctx, &command.API{Command: cmd.AppName, Arguments: cmd.AppArgs})
	if err != nil {
		return "", err
	}
	res, ok := NewResponse(raw).Get()
	if !ok {
		return res, fmt.Errorf("failed to execute api '%v': %v", cmd.AppName, res)
	}

	return res, nil
}

func (s *SocketClientImpl) BgApi(ctx context.Context, cmd *Command) (string, error) {
	raw, err := s.Conn.SendCommand(ctx, &command.API{Command: cmd.AppName, Arguments: cmd.AppArgs, Background: true})
	if err != nil {
		return "", err
	}
	res, ok := NewResponse(raw).Get()
	if !ok {
		return res, fmt.Errorf("failed to execute api '%v': %v", cmd.AppName, res)
	}

	return res, nil
}

func (s *SocketClientImpl) Originate(ctx context.Context, input *Originator) (string, error) {
	if input.Gateway == "" {
		return "", shared.RequireField("gateway")
	}

	if input.DNIS == "" {
		return "", shared.RequireField("DNIS")
	}

	if input.ANI == "" {
		input.ANI = "sofia"
	}

	if input.Variables == nil {
		input.Variables = make(map[string]interface{})
	}

	if input.Profile == "" {
		input.Profile = "external"
	}

	if input.Timeout == 0 {
		input.Timeout = 30 * time.Second
	}

	timeoutMillis := int32(input.Timeout / time.Millisecond)
	input.Variables["sip_contact_user"] = input.ANI
	input.Variables["originate_timeout"] = string(timeoutMillis)
	input.Variables["origination_caller_id_name"] = input.ANI
	input.Variables["origination_caller_id_number"] = input.ANI

	if input.AutoAnswer {
		input.Variables["sip_h_Answer"] = "auto"
	} else {
		input.Variables["sip_h_Answer"] = "manual"
	}
	if input.AllowReject {
		input.Variables["sip_h_Reject"] = "allow"
	} else {
		input.Variables["sip_h_Reject"] = "deny"
	}
	input.Variables["sip_h_Direction"] = string(input.Direction)
	bleg := eslgo.Leg{CallURL: fmt.Sprintf("&sleep(%v)", timeoutMillis)}

	vars := make(map[string]string)
	for k, v := range input.Variables {
		vars[k] = fmt.Sprintf("%v", v)
	}

	aleg := eslgo.Leg{CallURL: fmt.Sprintf("sofia/%v/%v@%v", input.Profile, input.DNIS, input.Gateway)}
	raw, err := s.Conn.OriginateCall(ctx, input.Background, aleg, bleg, vars)
	if err != nil {
		return "", err
	}

	res, ok := NewResponse(raw).Get()
	if !ok {
		return res, fmt.Errorf("failed to originate call: %v", res)
	}

	return res, nil
}
