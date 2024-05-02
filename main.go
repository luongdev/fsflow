package main

import (
	"context"
	"github.com/luongdev/fsflow/freeswitch"
	"log"
	"time"
)

var _ freeswitch.ServerEventHandler = &ServerEventHandlerImpl{}

type ServerEventHandlerImpl struct {
}

func (s *ServerEventHandlerImpl) OnSession(ctx context.Context, req *freeswitch.Request) {
	res, err := req.Client.Execute(ctx, &freeswitch.Command{
		AppName: "answer",
		Uid:     req.UniqueId,
	})

	select {
	case <-time.After(30 * time.Second):
		res, err = req.Client.Execute(ctx, &freeswitch.Command{
			AppName: "hangup",
			Uid:     req.UniqueId,
		})
	}

	if err != nil {
		panic(err)
	}

	log.Printf("Response: %v", res)
}

func main() {
	server, _, err := freeswitch.NewFreeswitchSocket(&freeswitch.Config{
		FsHost:           "10.8.0.1",
		FsPort:           65021,
		FsPassword:       "Simplefs!!",
		ServerListenPort: 65022,
		Timeout:          5 * time.Second,
	})

	if err != nil {
		panic(err)
	}

	server.SetEventHandler(&ServerEventHandlerImpl{})

	exit := make(chan any)

	<-exit
}
