package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/luongdev/fsflow/freeswitch"
	fsflow "github.com/luongdev/fsflow/workflow"
	"github.com/luongdev/fsflow/workflow/workflows"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/.gen/go/shared"
	"log"
	"time"
)

var _ freeswitch.ServerEventHandler = &ServerEventHandlerImpl{}

type ServerEventHandlerImpl struct {
	cadenceClient *workflowserviceclient.Interface
}

func (s *ServerEventHandlerImpl) OnSession(ctx context.Context, req *freeswitch.Request) {
	_, err := req.Client.Execute(ctx, &freeswitch.Command{
		AppName: "multiset",
		Uid:     req.UniqueId,
		AppArgs: fmt.Sprintf("api_after_bridge='uuid_transfer %v -bleg hold_call XML public'", req.UniqueId),
	})
	_, err = req.Client.Execute(ctx, &freeswitch.Command{AppName: "answer", Uid: req.UniqueId})
	_, err = req.Client.Execute(ctx, &freeswitch.Command{
		AppName: "sleep",
		AppArgs: fmt.Sprintf("%v", uint(time.Second*5/time.Millisecond)),
		Uid:     req.UniqueId,
	})

	initializer := req.GetHeader("variable_init")
	input, err := json.Marshal(workflows.InboundWorkflowInput{
		SessionId:   req.UniqueId,
		ANI:         req.ANI,
		DNIS:        req.DNIS,
		Initializer: initializer,
		Timeout:     60 * time.Second,
	})

	if err != nil {
		log.Printf("Failed to marshal input %v", err)
		return
	}

	domain := "default"
	taskList := "demo-task-list"
	workflowType := "workflows.InboundWorkflow"
	executionTimeout := int32(60)
	closeTimeout := int32(60)

	wfReq := &shared.StartWorkflowExecutionRequest{
		Domain:                              &domain,
		WorkflowId:                          &req.UniqueId,
		RequestId:                           &req.UniqueId,
		WorkflowType:                        &shared.WorkflowType{Name: &workflowType},
		TaskList:                            &shared.TaskList{Name: &taskList},
		Input:                               input,
		ExecutionStartToCloseTimeoutSeconds: &executionTimeout,
		TaskStartToCloseTimeoutSeconds:      &closeTimeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	resp, err := (*s.cadenceClient).StartWorkflowExecution(ctx, wfReq)
	log.Printf("successfully started InboundWorkflow workflow %+v", resp)

}

func main() {
	server, client, err := freeswitch.NewFreeswitchSocket(&freeswitch.Config{
		FsHost:           "10.8.0.1",
		FsPort:           65021,
		FsPassword:       "Simplefs!!",
		ServerListenPort: 65022,
		Timeout:          5 * time.Second,
	})

	if err != nil {
		panic(err)
	}

	w, err := fsflow.NewFreeswitchWorker(fsflow.Config{
		CadenceTaskList:   "demo-task-list",
		CadenceHost:       "103.141.141.60",
		CadenceClientName: "demo-client",
	}, &fsflow.FreeswitchWorkerOptions{Domain: "default", FsClient: &client})

	server.SetEventHandler(&ServerEventHandlerImpl{cadenceClient: w.CadenceClient})

	err = w.Start()
	if err != nil {
		panic(err)
	}

	exit := make(chan any)

	<-exit
}
