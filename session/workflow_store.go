package session

import (
	"fmt"
	"github.com/luongdev/fsflow/shared"
)

var _ Store = (*StoreImpl)(nil)

type Store interface {
	GetActivity(key string) (shared.FreeswitchActivity, error)
	SetActivity(key string, activity shared.FreeswitchActivity)

	GetWorkflow(key string) (shared.FreeswitchWorkflow, error)
	SetWorkflow(key string, workflow shared.FreeswitchWorkflow)
}

type StoreImpl struct {
	activities map[string]shared.FreeswitchActivity
	workflows  map[string]shared.FreeswitchWorkflow
}

func NewWorkflowStore() *StoreImpl {
	return &StoreImpl{
		activities: make(map[string]shared.FreeswitchActivity),
		workflows:  make(map[string]shared.FreeswitchWorkflow),
	}
}

func (s *StoreImpl) GetActivity(key string) (shared.FreeswitchActivity, error) {
	if a, ok := s.activities[key]; ok {
		return a, nil
	}

	return nil, fmt.Errorf("activity [%v] not found", key)
}

func (s *StoreImpl) SetActivity(key string, activity shared.FreeswitchActivity) {
	if _, ok := s.activities[key]; !ok {
		s.activities[key] = activity
	}
}

func (s *StoreImpl) GetWorkflow(key string) (shared.FreeswitchWorkflow, error) {
	if w, ok := s.workflows[key]; ok {
		return w, nil
	}

	return nil, fmt.Errorf("workflow [%v] not found", key)
}

func (s *StoreImpl) SetWorkflow(key string, workflow shared.FreeswitchWorkflow) {
	if _, ok := s.workflows[key]; !ok {
		s.workflows[key] = workflow
	}
}
