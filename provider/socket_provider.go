package provider

import "github.com/luongdev/fsflow/freeswitch"

var _ SocketProvider = (*SocketProviderImpl)(nil)

type SocketProviderImpl struct {
	store *freeswitch.SocketStore
}

func NewSocketProvider(store *freeswitch.SocketStore) SocketProviderImpl {
	return SocketProviderImpl{store: store}
}

func (s *SocketProviderImpl) GetClient(key string) freeswitch.SocketClient {
	c, err := (*s.store).Get(key)
	if err != nil {
		c, err = (*s.store).Get(freeswitch.DefaultClient)
		if err != nil {
			return nil
		}
	}

	return c
}
