package main

import (
	"context"
	"encoding/json"
	"github.com/luongdev/fsflow/freeswitch"
	fsflow "github.com/luongdev/fsflow/workflow"
	"github.com/luongdev/fsflow/workflow/activities"
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
		AppName: "answer",
		Uid:     req.UniqueId,
	})

	sessionId := req.UniqueId
	phoneNumber := req.GetHeader("Channel-Caller-ID-Number")
	if phoneNumber == "" {
		phoneNumber = req.GetHeader("Channel-ANI")
	}
	dialedNumber := req.GetHeader("variable_sip_to_user")
	if dialedNumber == "" {
		dialedNumber = req.GetHeader("variable_sip_req_user")
	}

	initializer := req.GetHeader("variable_init")
	if initializer == "" {
		initializer = "https://omicx.vn"
	}

	log.Printf("session_id: %s, phone_number: %s, dialed_number: %s", sessionId, phoneNumber, dialedNumber)

	input, err := json.Marshal(workflows.InboundWorkflowInput{
		SessionId:    sessionId,
		PhoneNumber:  phoneNumber,
		DialedNumber: dialedNumber,
		Initializer:  initializer,
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
		WorkflowId:                          &sessionId,
		RequestId:                           &sessionId,
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

	fw := w.(*fsflow.FreeswitchWorker)

	ibWorkflow := workflows.NewInboundWorkflow(&client)
	fw.AddWorkflow(ibWorkflow)

	hangupActivity := activities.NewHangupActivity(&client)
	fw.AddActivity(hangupActivity)

	server.SetEventHandler(&ServerEventHandlerImpl{cadenceClient: fw.CadenceClient})

	err = w.Start()
	if err != nil {
		panic(err)
	}

	exit := make(chan any)

	<-exit
}
