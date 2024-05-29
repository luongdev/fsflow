package freeswitch

import (
	"context"
	"fmt"
	error2 "github.com/luongdev/fsflow/errors"
	"github.com/percipia/eslgo"
	"github.com/percipia/eslgo/command"
	"github.com/percipia/eslgo/command/call"
	"strings"
	"time"
)

var _ SocketClient = (*SocketClientImpl)(nil)

type Filter struct {
}

func (f *Filter) BuildMessage() string {
	return "filter "
}

type SocketClientImpl struct {
	*eslgo.Conn
}

func NewSocketClient(conn *eslgo.Conn) SocketClientImpl {
	return SocketClientImpl{Conn: conn}
}

func (s *SocketClientImpl) AllEvents(ctx context.Context) error {
	raw, err := s.Conn.SendCommand(ctx, &command.Event{
		Format: "plain",
		Listen: []string{"ALL"},
	})

	if err != nil {
		return err
	}
	res, ok := NewResponse(raw).Get()
	if !ok {
		return fmt.Errorf("failed to listen to all events: %v", res)
	}

	return nil
}

func (s *SocketClientImpl) MyEvents(ctx context.Context, id string) error {
	raw, err := s.Conn.SendCommand(ctx, &command.MyEvents{Format: "plain", UUID: id})

	if err != nil {
		return err
	}
	res, ok := NewResponse(raw).Get()
	if !ok {
		return fmt.Errorf("failed to listen to  myevents: %v", res)
	}

	return nil
}

func (s *SocketClientImpl) AddFilter(ctx context.Context, header, value string) error {
	raw, err := s.Conn.SendCommand(ctx, &command.Filter{
		EventHeader: header,
		FilterValue: value,
		Delete:      false,
	})

	if err != nil {
		return err
	}
	res, ok := NewResponse(raw).Get()
	if !ok {
		return fmt.Errorf("failed to add filter events: %v", res)
	}

	return nil
}

func (s *SocketClientImpl) DelFilter(ctx context.Context, header, value string) error {
	raw, err := s.Conn.SendCommand(ctx, &command.Filter{
		EventHeader: header,
		FilterValue: value,
		Delete:      true,
	})

	if err != nil {
		return err
	}
	res, ok := NewResponse(raw).Get()
	if !ok {
		return fmt.Errorf("failed to delete filter events: %v", res)
	}

	return nil
}

func (s *SocketClientImpl) EventListener(id string, listener EventListener) string {
	if listener == nil {
		return ""
	}
	return s.Conn.RegisterEventListener(id, func(event *eslgo.Event) {
		listener(NewEvent(s, event))
	})
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
	uid, err := input.GetUIdOrDefault()
	if err != nil {
		return "", err
	}

	if input.Gateway == "" {
		return "", error2.RequireField("gateway")
	}

	if input.OrigTo == "" {
		return "", error2.RequireField("origTo")
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
	input.Variables["sip_contact_user"] = input.OrigFrom
	input.Variables["originate_timeout"] = string(timeoutMillis)
	input.Variables["origination_caller_id_name"] = input.OrigFrom
	input.Variables["origination_caller_id_number"] = input.OrigFrom
	input.Variables["Direction"] = string(input.Direction)

	input.Variables["sid"] = input.SessionId
	input.Variables["X-Session-ID"] = input.SessionId
	input.Variables["disable_q850_reason"] = true

	if input.AutoAnswer {
		input.Variables["X-Answer"] = "auto"
	} else {
		input.Variables["X-Answer"] = "manual"
	}
	if input.AllowReject {
		input.Variables["X-Reject"] = "allow"
	} else {
		input.Variables["X-Reject"] = "deny"
	}
	input.Variables["X-Direction"] = string(input.Direction)
	var bleg eslgo.Leg

	if input.Extension != "" {
		bleg = eslgo.Leg{CallURL: fmt.Sprintf("%v", input.Extension)}
	} else {
		bleg = eslgo.Leg{CallURL: fmt.Sprintf("&sleep(%v)", timeoutMillis)}
	}

	vars := make(map[string]string)
	for k, v := range input.Variables {
		if strings.HasPrefix(k, "X-") {
			k = "sip_h_" + k
		}
		vars[k] = fmt.Sprintf("%v", v)
	}

	aleg := eslgo.Leg{
		CallURL:      fmt.Sprintf("sofia/%v/%v@%v", input.Profile, input.OrigTo, input.Gateway),
		LegVariables: map[string]string{"origination_uuid": fmt.Sprintf("%v", uid)},
	}
	raw, err := s.Conn.OriginateCall(ctx, input.Background, aleg, bleg, vars)
	if err != nil {
		return "", err
	}

	res, ok := NewResponse(raw).Get()
	if !ok && res == "" {
		return res, fmt.Errorf("failed to originate call: %v", res)
	}

	return res, nil
}

func (s *SocketClientImpl) SendEvent(ctx context.Context, cmd *Command) (string, error) {
	raw, err := s.Conn.SendCommand(ctx, &command.SendEvent{
		Name: "CUSTOM",
		Headers: map[string][]string{
			"Event-Subclass": {"callmanager::event"},
			"Session-Id":     {cmd.Uid},
		},
	})

	if err != nil {
		return "", err
	}

	res, ok := NewResponse(raw).Get()
	if !ok {
		return res, fmt.Errorf("failed to originate call: %v", res)
	}

	return res, nil
}
