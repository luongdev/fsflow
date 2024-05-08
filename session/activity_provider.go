package session

import (
	"github.com/luongdev/fsflow/shared"
)

type ActivityProvider interface {
	GetActivity(key string) shared.FreeswitchActivity
}

var _ ActivityProvider = (*ActivityProviderImpl)(nil)

type ActivityProviderImpl struct {
	store Store
}

func NewActivityProvider(store Store) *ActivityProviderImpl {
	return &ActivityProviderImpl{store: store}
}

func (a *ActivityProviderImpl) GetActivity(key string) shared.FreeswitchActivity {
	activity, err := a.store.GetActivity(key)
	if err != nil {
		return nil
	}

	return activity
}
