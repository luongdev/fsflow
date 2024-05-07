package workflow

import (
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/luongdev/fsflow/shared"
	"github.com/luongdev/fsflow/workflow/activities"
	"github.com/luongdev/fsflow/workflow/workflows"
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
	SocketProvider freeswitch.SocketProvider
}

type FreeswitchWorker struct {
	worker.Worker
	socketProvider freeswitch.SocketProvider
	CadenceClient  *workflowserviceclient.Interface

	workflows  []shared.FreeswitchWorkflow
	activities []shared.FreeswitchActivity
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
		Worker:         w,
		CadenceClient:  &client,
		socketProvider: opts.SocketProvider,
		workflows:      make([]shared.FreeswitchWorkflow, 0),
		activities:     make([]shared.FreeswitchActivity, 0),
	}

	fsWorker.AddWorkflow(workflows.NewInboundWorkflow(opts.SocketProvider))

	fsWorker.AddActivity(activities.NewEventActivity(opts.SocketProvider))
	fsWorker.AddActivity(activities.NewBridgeActivity(opts.SocketProvider))
	fsWorker.AddActivity(activities.NewHangupActivity(opts.SocketProvider))
	fsWorker.AddActivity(activities.NewOriginateActivity(opts.SocketProvider))
	fsWorker.AddActivity(activities.NewSessionInitActivity(opts.SocketProvider))

	return fsWorker, nil
}

func (w *FreeswitchWorker) Start() error {
	for _, ww := range w.workflows {
		w.Worker.RegisterWorkflowWithOptions(ww.Handler(), workflow.RegisterOptions{Name: ww.Name()})
	}

	for _, wa := range w.activities {
		w.Worker.RegisterActivityWithOptions(wa.Handler(), activity.RegisterOptions{Name: wa.Name()})
	}

	err := w.Worker.Start()
	if err != nil {
		return err
	}

	return nil
}

func (w *FreeswitchWorker) AddWorkflow(workflow shared.FreeswitchWorkflow) {
	w.workflows = append(w.workflows, workflow)
}

func (w *FreeswitchWorker) AddActivity(activity shared.FreeswitchActivity) {
	w.activities = append(w.activities, activity)
}
