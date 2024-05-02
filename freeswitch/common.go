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

type Request struct {
	*eslgo.RawResponse
	UniqueId string
	ANI      string
	DNIS     string
	Domain   string
	Client   SocketClient
}

func NewRequest(conn *eslgo.Conn, raw *eslgo.RawResponse) *Request {
	r := &Request{
		Client:      &SocketClientImpl{conn},
		RawResponse: raw,
	}

	r.UniqueId = r.getUniqueId()

	r.ANI = r.GetHeader("Channel-Caller-ID-Number")
	if r.ANI == "" {
		r.ANI = r.GetHeader("Channel-ANI")
	}

	r.DNIS = r.GetHeader("variable_sip_to_user")
	if r.DNIS == "" {
		r.DNIS = r.GetHeader("variable_sip_req_user")
	}

	return r
}

func (r *Request) getUniqueId() string {
	if r.HasHeader("Channel-Call-UUID") {
		return r.GetHeader("Channel-Call-UUID")
	}

	if r.HasHeader("Unique-ID") {
		return r.GetHeader("Unique-ID")
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
	BridgeTo    string
	Direction   Direction
	ANI         string
	DNIS        string
	Gateway     string
	Profile     string
	Timeout     time.Duration
	Variables   map[string]interface{}
}

type ServerEventHandler interface {
	OnSession(ctx context.Context, req *Request)
}

type SocketClient interface {
	Execute(ctx context.Context, cmd *Command) (string, error)
	Originate(ctx context.Context, o *Originator) (string, error)
	Api(ctx context.Context, cmd *Command) (string, error)
	BgApi(ctx context.Context, cmd *Command) (string, error)
	Close()
}

type SocketServer interface {
	ListenAndServe() error
	SetEventHandler(handler ServerEventHandler)
}

func removeUnwantedChars(s string) string {
	return strings.TrimRight(strings.TrimLeft(s, " \t"), "\n")
}
