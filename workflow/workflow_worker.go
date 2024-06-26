package workflow

import (
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/session"
	"github.com/luongdev/fsflow/session/activities"
	"github.com/luongdev/fsflow/shared"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"io"
	"os"
	"time"
)

type FreeswitchWorkerOptions struct {
	Domain         string
	WorkflowStore  session.Store
	SocketProvider freeswitch.SocketProvider
}

type FreeswitchWorker struct {
	worker         worker.Worker
	socketProvider freeswitch.SocketProvider
	CadenceClient  *workflowserviceclient.Interface

	workflows  map[string]bool
	activities map[string]bool

	store session.Store
}

func NewFreeswitchWorker(c *Config, opts *FreeswitchWorkerOptions) (*FreeswitchWorker, error) {
	client, err := NewCadenceClient(c)
	if err != nil {
		return nil, err
	}

	logger, err := NewLogger()
	if err != nil {
		return nil, err
	}

	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Prefix:   c.ClientName,
		Tags:     map[string]string{"env": os.Getenv("ENV")},
		Reporter: tally.NullStatsReporter,
	}, 5*time.Second)

	defer func(closer io.Closer) {
		_ = closer.Close()
	}(closer)

	workerOptions := worker.Options{Logger: logger, MetricsScope: scope}
	w := worker.New(client, opts.Domain, c.TaskList, workerOptions)

	fsWorker := &FreeswitchWorker{
		worker:         w,
		CadenceClient:  &client,
		socketProvider: opts.SocketProvider,
		workflows:      make(map[string]bool),
		activities:     make(map[string]bool),
		store:          opts.WorkflowStore,
	}

	fsWorker.AddActivity(activities.NewCallbackActivity())
	fsWorker.AddActivity(activities.NewSessionInitActivity())
	fsWorker.AddActivity(activities.NewSetActivity(opts.SocketProvider))
	fsWorker.AddActivity(activities.NewGetActivity(opts.SocketProvider))
	fsWorker.AddActivity(activities.NewEventActivity(opts.SocketProvider))
	fsWorker.AddActivity(activities.NewBridgeActivity(opts.SocketProvider))
	fsWorker.AddActivity(activities.NewHangupActivity(opts.SocketProvider))
	fsWorker.AddActivity(activities.NewOriginateActivity(opts.SocketProvider))

	return fsWorker, nil
}

func (w *FreeswitchWorker) Start() error {
	for wName := range w.workflows {
		ww, err := w.store.GetWorkflow(wName)
		if err != nil {
			return err
		}
		w.worker.RegisterWorkflowWithOptions(ww.Handler(), workflow.RegisterOptions{Name: ww.Name()})
	}

	for aName := range w.activities {
		wa, err := w.store.GetActivity(aName)
		if err != nil {
			return err
		}
		w.worker.RegisterActivityWithOptions(wa.Handler(), activity.RegisterOptions{Name: wa.Name()})
	}

	err := w.worker.Start()
	if err != nil {
		return err
	}

	return nil
}

func (w *FreeswitchWorker) AddWorkflow(workflow shared.FreeswitchWorkflow) {
	if workflow != nil {
		w.workflows[workflow.Name()] = true
		w.store.SetWorkflow(workflow.Name(), workflow)
	}
}

func (w *FreeswitchWorker) AddActivity(activity shared.FreeswitchActivity) {
	if activity != nil {
		w.activities[activity.Name()] = true
		w.store.SetActivity(activity.Name(), activity)
	}
}

func (w *FreeswitchWorker) GetStore() session.Store {
	return w.store
}
