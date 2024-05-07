package freeswitch

import "fmt"

var _ SocketStore = (*SocketStoreImpl)(nil)

const DefaultClient = "default"

type SocketStoreImpl struct {
	clients map[string]*SocketClient
}

func NewSocketStore() *SocketStoreImpl {
	return &SocketStoreImpl{
		clients: make(map[string]*SocketClient),
	}
}

func (s *SocketStoreImpl) Set(key string, client SocketClient) {
	if _, ok := s.clients[key]; ok {
		return
	}

	s.clients[key] = &client
}

func (s *SocketStoreImpl) Get(key string) (SocketClient, error) {
	if c, ok := s.clients[key]; ok {
		return *c, nil
	}

	return nil, fmt.Errorf("client [%v] not found", key)
}

func (s *SocketStoreImpl) Del(key string) error {
	if _, ok := s.clients[key]; ok {
		delete(s.clients, key)
		return nil
	}

	return fmt.Errorf("client [%v] not found", key)
}
