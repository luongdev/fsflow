package freeswitch

var _ SocketProvider = (*SocketProviderImpl)(nil)

type SocketProviderImpl struct {
	store *SocketStore
}

func NewSocketProvider(store *SocketStore) *SocketProviderImpl {
	return &SocketProviderImpl{store: store}
}

func (s *SocketProviderImpl) GetClient(key string) SocketClient {
	c, err := (*s.store).Get(key)
	if err != nil {
		c, err = (*s.store).Get(DefaultClient)
		if err != nil {
			return nil
		}
	}

	return c
}
