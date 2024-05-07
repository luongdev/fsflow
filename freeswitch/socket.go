package freeswitch

import (
	"context"
	"fmt"
	"github.com/percipia/eslgo"
	"time"
)

func NewFreeswitchSocket(c *Config) (SocketServer, SocketClient, error) {
	if c.Port == 0 {
		c.Port = 8021
	}

	if c.Host == "" {
		c.Host = "127.0.0.1"
	}

	if c.Password == "" {
		c.Password = "ClueCon"
	}

	if c.ListenOn == 0 {
		c.ListenOn = 65022
	}

	if c.Timeout == 0 {
		c.Timeout = 10 * time.Second
	}

	store := NewSocketStore()
	server := NewSocketServer(c.ListenOn, store)

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

	hostPort := fmt.Sprintf("%v:%v", c.Host, c.Port)
	conn, err := eslgo.Dial(hostPort, c.Password, func() {
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

	store.Set(DefaultClient, &client)

	return &server, &client, nil
}
