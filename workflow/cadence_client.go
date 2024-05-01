package workflow

import (
	"fmt"
	"github.com/luongdev/fsflow/errors"
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/zap"
	"os"
)

func NewCadenceClient(c Config) (workflowserviceclient.Interface, error) {
	if c.CadenceService == "" {
		c.CadenceService = "cadence-frontend"
	}
	if c.CadenceHost == "" {
		c.CadenceHost = "127.0.0.1"
	}
	if c.CadencePort == 0 {
		c.CadencePort = 7833
	}
	if c.CadenceClientName == "" {
		return nil, errors.RequireField("cadenceClientName")
	}
	if c.CadenceTaskList == "" {
		return nil, errors.RequireField("cadenceTaskList")
	}

	hostPort := fmt.Sprintf("%v:%v", c.CadenceHost, c.CadencePort)
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: c.CadenceClientName,
		Outbounds: yarpc.Outbounds{
			c.CadenceService: {Unary: grpc.NewTransport().NewSingleOutbound(hostPort)},
		},
	})
	err := dispatcher.Start()
	if err != nil {
		return nil, err
	}

	clientConfig := dispatcher.ClientConfig(c.CadenceService)
	itf := compatibility.NewThrift2ProtoAdapter(
		apiv1.NewDomainAPIYARPCClient(clientConfig),
		apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		apiv1.NewWorkerAPIYARPCClient(clientConfig),
		apiv1.NewVisibilityAPIYARPCClient(clientConfig),
	)

	return itf, nil
}

func NewLogger() (*zap.Logger, error) {
	var cfg zap.Config
	if os.Getenv("ENV") == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
