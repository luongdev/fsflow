package workflow

import (
	"github.com/luongdev/fsflow/freeswitch"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/worker"
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

	workflows  []interface{}
	activities []interface{}
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
		workflows:     make([]interface{}, 0),
		activities:    make([]interface{}, 0),
	}, nil
}

func (w *FreeswitchWorker) Start() error {
	for _, workflow := range w.workflows {
		w.Worker.RegisterWorkflow(workflow)
	}

	for _, activity := range w.activities {
		w.Worker.RegisterActivity(activity)
	}

	err := w.Worker.Start()
	if err != nil {
		return err
	}

	return nil
}

func (w *FreeswitchWorker) RegisterWorkflow(workflow interface{}) {
	w.workflows = append(w.workflows, workflow)
}

func (w *FreeswitchWorker) RegisterActivity(activity interface{}) {
	w.workflows = append(w.workflows, activity)
}
