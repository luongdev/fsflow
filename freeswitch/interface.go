package freeswitch

type SocketProvider interface {
	GetClient(key string) SocketClient
}
