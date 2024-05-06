package freeswitch

import (
	"context"
	"fmt"
	"github.com/percipia/eslgo"
)

var _ SocketServer = (*SocketServerImpl)(nil)

type SocketServerImpl struct {
	port               uint16
	serverEventHandler ServerEventHandler
	beforeSessionClose func()
}

func (s *SocketServerImpl) BeforeSessionClose(f func()) {
	if f != nil {
		s.beforeSessionClose = f
	}
}

func NewSocketServer(port uint16) *SocketServerImpl {
	return &SocketServerImpl{
		port: port,
	}
}

func (s *SocketServerImpl) SetEventHandler(handler ServerEventHandler) {
	if handler != nil {
		s.serverEventHandler = handler
	}
}

func (s *SocketServerImpl) ListenAndServe() error {
	listenAddr := fmt.Sprintf("0.0.0.0:%v", s.port)
	err := eslgo.ListenAndServe(listenAddr, func(ctx context.Context, conn *eslgo.Conn, connectResponse *eslgo.RawResponse) {
		if s.serverEventHandler != nil {
			client := NewSocketClient(conn)

			s.serverEventHandler.OnSession(ctx, NewRequest(client, connectResponse))

			if s.beforeSessionClose != nil {
				s.beforeSessionClose()
			}
		}
	})

	return err
}
