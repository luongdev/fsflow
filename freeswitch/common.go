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
		ANI:         getANI(raw),
		DNIS:        getDNIS(raw),
		Domain:      getDomain(raw),
		UniqueId:    getUniqueId(raw),
	}

	return r
}

func getDomain(raw *eslgo.RawResponse) string {
	if raw != nil {
		if raw.HasHeader("variable_domain") {
			return raw.GetHeader("variable_domain")
		}
	}

	return ""
}

func getANI(raw *eslgo.RawResponse) string {
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

func getDNIS(raw *eslgo.RawResponse) string {
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

func getUniqueId(raw *eslgo.RawResponse) string {
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

type Command struct {
	AppName string `json:"appName"`
	AppArgs string `json:"appArgs"`
	Uid     string `json:"uid"`
}

type Originator struct {
	AutoAnswer  bool
	AllowReject bool
	Background  bool
	Direction   Direction
	ANI         string
	DNIS        string
	Gateway     string
	Profile     string
	Timeout     time.Duration
	Extension   string
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
