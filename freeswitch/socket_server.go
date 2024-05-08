package freeswitch

import (
	"context"
	"fmt"
	"github.com/percipia/eslgo"
	"log"
)

var _ SocketServer = (*SocketServerImpl)(nil)

type SocketServerImpl struct {
	port               uint16
	serverEventHandler ServerEventHandler
	sessionClosed      func(sid string)
	store              SocketStore
}

func (s *SocketServerImpl) Store() *SocketStore {
	return &s.store
}

func (s *SocketServerImpl) OnSessionClosed(f func(sid string)) {
	if f != nil {
		s.sessionClosed = f
	}
}

func NewSocketServer(port uint16, store SocketStore) SocketServerImpl {
	return SocketServerImpl{
		port:  port,
		store: store,
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
		client := NewSocketClient(conn)
		req := NewRequest(&client, connectResponse)
		_, _ = client.Execute(ctx, &Command{
			AppName: "multiset",
			Uid:     req.UniqueId,
			AppArgs: fmt.Sprintf("park_after_bridge=true session_id=%v", req.UniqueId),
		})

		go func() {
			res, err := client.Execute(ctx, &Command{AppName: "answer", Uid: req.UniqueId})
			if err != nil {
				log.Printf("Failed to answer call %v", err)
				return
			}
			log.Printf("Answered call %v: %v", req.UniqueId, res)
		}()

		s.store.Set(req.UniqueId, &client)
		if s.serverEventHandler != nil {
			go s.serverEventHandler.OnSession(ctx, req)
		}

		client.AllEvents(ctx)
		client.AddFilter(ctx, "variable_session_id", req.UniqueId)

		client.EventListener("ALL", func(event *Event) {
			if event.SessionId != "" {
				if event.UniqueId == event.SessionId {
					go s.serverEventHandler.OnAlegEvent(ctx, event)
				} else {
					go s.serverEventHandler.OnBlegEvent(ctx, event)
				}
			}
			go s.serverEventHandler.OnEvent(ctx, event)

			go func() {
				if r := recover(); r != nil {
					log.Fatalf("Recovered from panic: %v", r)
				}
			}()
		})

		select {
		case <-ctx.Done():
			if s.sessionClosed != nil {
				s.sessionClosed(req.UniqueId)
			}

			err := s.store.Del(req.UniqueId)
			if err != nil {
				log.Fatalf("Failed to delete session %v: %v", req.UniqueId, err)
			}
			break
		}

		log.Printf("Outbound connection for session %v closed", req.UniqueId)
	})

	return err
}
