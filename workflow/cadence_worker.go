package workflow

import (
	"github.com/luongdev/fsflow/freeswitch"
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
	FsClient *freeswitch.SocketClient
	Domain   string
}

type FreeswitchWorker struct {
	worker.Worker
	config        Config
	fsClient      *freeswitch.SocketClient
	CadenceClient *workflowserviceclient.Interface

	workflows  []shared.FreeswitchWorkflow
	activities []shared.FreeswitchActivity
}

func NewFreeswitchWorker(c Config, opts *FreeswitchWorkerOptions) (worker.Worker, error) {
	client, err := NewCadenceClient(c)
	if err != nil {
		return nil, err
	}

	logger, err := NewLogger()
	if err != nil {
		return nil, err
	}

	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Prefix:   c.CadenceClientName,
		Tags:     map[string]string{"env": os.Getenv("ENV")},
		Reporter: tally.NullStatsReporter,
	}, 5*time.Second)

	defer func(closer io.Closer) {
		_ = closer.Close()
	}(closer)

	workerOptions := worker.Options{Logger: logger, MetricsScope: scope}
	w := worker.New(client, opts.Domain, c.CadenceTaskList, workerOptions)

	return &FreeswitchWorker{
		Worker:        w,
		CadenceClient: &client,
		fsClient:      opts.FsClient,
		workflows:     make([]shared.FreeswitchWorkflow, 0),
		activities:    make([]shared.FreeswitchActivity, 0),
	}, nil
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
