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

func NewCadenceClient(c *Config) (workflowserviceclient.Interface, error) {
	if c.ServiceName == "" {
		c.ServiceName = "cadence-frontend"
	}
	if c.Host == "" {
		c.Host = "127.0.0.1"
	}
	if c.Port == 0 {
		c.Port = 7833
	}
	if c.ClientName == "" {
		return nil, errors.RequireField("cadenceClientName")
	}
	if c.TaskList == "" {
		return nil, errors.RequireField("cadenceTaskList")
	}

	hostPort := fmt.Sprintf("%v:%v", c.Host, c.Port)
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: c.ClientName,
		Outbounds: yarpc.Outbounds{
			c.ServiceName: {Unary: grpc.NewTransport().NewSingleOutbound(hostPort)},
		},
	})
	err := dispatcher.Start()
	if err != nil {
		return nil, err
	}

	clientConfig := dispatcher.ClientConfig(c.ServiceName)
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
