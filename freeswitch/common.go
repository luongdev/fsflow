package freeswitch

import (
	"context"
	"github.com/percipia/eslgo"
	"strings"
	"time"
)

type Direction string

const (
	Inbound  Direction = "inbound"
	Outbound Direction = "outbound"
)

type Status string

const (
	Success Status = "+OK"
	Failure Status = "-ERR"
	Syntax  Status = "-USAGE"
)

type Response struct {
	*eslgo.RawResponse
}

func NewResponse(res *eslgo.RawResponse) *Response {
	return &Response{RawResponse: res}
}

func (c *Response) Get() (string, bool) {
	var body string
	if c.Body == nil {
		if c.HasHeader("Reply-Text") {
			body = c.GetHeader("Reply-Text")
		}
		if body == "" {
			return "", false
		}
	} else {
		body = string(c.Body)
	}

	res, found := strings.CutPrefix(body, string(Failure))
	if found {
		return strings.TrimSpace(res), false
	}

	found = strings.HasPrefix(body, string(Syntax))
	if found {
		return strings.TrimSpace(body), false
	}

	res, found = strings.CutPrefix(body, string(Success))
	if found {
		return removeUnwantedChars(res), true
	}

	return removeUnwantedChars(res), true
}

type Event struct {
	*eslgo.Event
	UniqueId  string
	SessionId string
	Domain    string
	Client    SocketClient
}

func NewEvent(client SocketClient, event *eslgo.Event) *Event {
	e := &Event{
		Event:     event,
		Client:    client,
		UniqueId:  getUniqueId(event),
		Domain:    getDomain(event),
		SessionId: getSessionId(event),
	}

	return e
}

type Request struct {
	*eslgo.RawResponse
	UniqueId  string
	SessionId string
	ANI       string
	DNIS      string
	Domain    string
	Client    SocketClient
}

func NewRequest(client SocketClient, raw *eslgo.RawResponse) *Request {
	r := &Request{
		RawResponse: raw,
		Client:      client,
		ANI:         getANI(raw),
		DNIS:        getDNIS(raw),
		Domain:      getDomain(raw),
		UniqueId:    getUniqueId(raw),
	}

	sid := getSessionId(raw)
	if sid == "" {
		sid = r.UniqueId
	}
	r.SessionId = sid

	return r
}

type RawData interface {
	HasHeader(header string) bool
	GetHeader(header string) string
}

func getDomain(raw RawData) string {
	if raw != nil {
		if raw.HasHeader("variable_domain") {
			return raw.GetHeader("variable_domain")
		}
	}

	return ""
}

func getANI(raw RawData) string {
	if raw != nil {
		if raw.HasHeader("Channel-Caller-ID-Number") {
			return raw.GetHeader("Channel-Caller-ID-Number")
		}

		if raw.HasHeader("Channel-ANI") {
			return raw.GetHeader("Channel-ANI")
		}
	}

	return ""
}

func getDNIS(raw RawData) string {
	if raw != nil {
		if raw.HasHeader("variable_sip_to_user") {
			return raw.GetHeader("variable_sip_to_user")
		}

		if raw.HasHeader("variable_sip_req_user") {
			return raw.GetHeader("variable_sip_req_user")
		}
	}

	return ""
}

func getUniqueId(raw RawData) string {
	if raw != nil {
		if raw.HasHeader("Channel-Call-UUID") {
			return raw.GetHeader("Channel-Call-UUID")
		}

		if raw.HasHeader("Unique-ID") {
			return raw.GetHeader("Unique-ID")
		}
	}

	return ""
}

func getSessionId(raw RawData) string {
	if raw != nil {
		if raw.HasHeader("variable_sid") {
			return raw.GetHeader("variable_sid")
		}

		if raw.HasHeader("variable_sip_h_X-Session-ID") {
			return raw.GetHeader("variable_sip_h_X-Session-ID")
		}
	}

	return ""
}

type Command struct {
	AppName string `json:"appName"`
	AppArgs string `json:"appArgs"`
	Uid     string `json:"uid"`
}

type Originator struct {
	AutoAnswer  bool
	AllowReject bool
	Background  bool
	Callback    string
	Direction   Direction
	ANI         string
	DNIS        string
	Gateway     string
	Profile     string
	Timeout     time.Duration
	Extension   string
	SessionId   string
	Variables   map[string]interface{}
}

type EventListener func(req *Event)

type ServerEventHandler interface {
	OnSession(ctx context.Context, req *Request)

	OnEvent(ctx context.Context, event *Event)
	OnAlegEvent(ctx context.Context, event *Event)
	OnBlegEvent(ctx context.Context, event *Event)
}

type SocketClient interface {
	Execute(ctx context.Context, cmd *Command) (string, error)
	Originate(ctx context.Context, o *Originator) (string, error)
	Api(ctx context.Context, cmd *Command) (string, error)
	BgApi(ctx context.Context, cmd *Command) (string, error)
	AllEvents(ctx context.Context) error
	MyEvents(ctx context.Context, id string) error
	EventListener(id string, listener EventListener) string
	SendEvent(ctx context.Context, cmd *Command) (string, error)
	AddFilter(ctx context.Context, header, value string) error
	DelFilter(ctx context.Context, header, value string) error
	Close()
}

type SocketServer interface {
	Store() *SocketStore
	ListenAndServe() error
	SetEventHandler(handler ServerEventHandler)
	OnSessionClosed(func(sid string))
}

type SocketStore interface {
	Set(key string, client SocketClient)
	Get(key string) (SocketClient, error)
	Del(key string) error
}

func removeUnwantedChars(s string) string {
	return strings.TrimRight(strings.TrimLeft(s, " \t"), "\n")
}
