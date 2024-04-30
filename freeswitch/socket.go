package freeswitch

import (
	"context"
	"fmt"
	"github.com/percipia/eslgo"
	"time"
)

func NewFreeswitchSocket(c *Config) (SocketServer, SocketClient, error) {
	if c.FsPort == 0 {
		c.FsPort = 8021
	}

	if c.FsHost == "" {
		c.FsHost = "127.0.0.1"
	}

	if c.FsPassword == "" {
		c.FsPassword = "ClueCon"
	}

	if c.ServerListenPort == 0 {
		c.ServerListenPort = 65022
	}

	server := NewSocketServer(c.ServerListenPort)

	panicChan := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		panicChan <- err
	}()

	select {
	case err := <-panicChan:
		return nil, nil, err
	case <-time.After(c.Timeout):
		break
	}

	hostPort := fmt.Sprintf("%v:%v", c.FsHost, c.FsPort)
	conn, err := eslgo.Dial(hostPort, c.FsPassword, func() {
		fmt.Printf("Server %v disconnected", hostPort)
	})

	if err != nil {
		return nil, nil, err
	}

	client := NewSocketClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	_, err = client.Api(ctx, &Command{
		AppName: "show",
		AppArgs: "status",
	})

	if err != nil {
		return nil, nil, err
	}

	return server, client, nil
}
