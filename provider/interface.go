package provider

import "github.com/luongdev/fsflow/freeswitch"

type SocketProvider interface {
	GetClient(key string) freeswitch.SocketClient
}
