package client_service

import (
	"context"
	"time"

	"github.com/CGSG-2021-AE4/rc2/api"
	tcpw "github.com/CGSG-2021-AE4/rc2/pkg/tcp_wrapper"
)

// Client connection
type Conn struct {
	cs    *Service
	Conn  *tcpw.Conn
	Login string

	runCtx  context.Context
	stopRun context.CancelFunc
}

func NewClient(cs *Service, login string, c *tcpw.Conn) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	return &Conn{
		cs:      cs,
		Conn:    c,
		Login:   login,
		runCtx:  ctx,
		stopRun: cancel,
	}
}

func (c *Conn) RunSync() error {
	// Now it will only start it and... wait until end
	// It won't listen because I need to somehow get answer on other requests
	<-c.runCtx.Done()

	// Send close message and close
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Microsecond)
	defer cancel()
	if err := c.Conn.Write(ctx, tcpw.MsgTypeClose, []byte{}); err != nil {
		return err
	}
	return c.Conn.Close()
}

func (c *Conn) Run() {
	go func() {
		api.RunAndLog(c.RunSync, "client "+c.Login)
	}()
}

func (c *Conn) Close() {
	c.stopRun() // Cancel context
}
