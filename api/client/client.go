package client

import (
	"context"
	"encoding/json"
	"log"
	"os/exec"
	"path"
	"time"

	"github.com/CGSG-2021-AE4/rc2/api"
	tcpw "github.com/CGSG-2021-AE4/rc2/pkg/tcp_wrapper"
)

const (
	connectTimeout = 3 * time.Second
)

type Conn struct {
	config *Config

	conn *tcpw.Conn

	runCtx  context.Context
	stopRun context.CancelFunc
}

func NewConn(config *Config) (*Conn, error) {

	ctx, cancel := context.WithCancel(context.Background())
	return &Conn{
		config: config,
		conn:   nil,

		runCtx:  ctx,
		stopRun: cancel,
	}, nil
}

func (c *Conn) startScript(scriptName string, query string) error {
	for i := range len(c.config.Scripts) {
		if c.config.Scripts[i].Name == scriptName {
			filepath := c.config.Scripts[i].File
			if !path.IsAbs(filepath) {
				filepath = path.Join(path.Dir(c.config.HomeDir), filepath)
			}
			cmd := exec.Command(filepath + " " + query)
			if err := cmd.Start(); err != nil {
				return err
			}
			return nil
		}
	}
	return api.Error("Script not found")
}

func (c *Conn) handleRequest(buf []byte) error {
	// Decode message
	var rawMsg api.MainLoopMsg
	if err := json.Unmarshal(buf, &rawMsg); err != nil {
		return err
	}

	// Check password
	if rawMsg.Password != c.config.Password {
		return api.Error("Wrong password")
	}

	switch rawMsg.Type {
	case "script":
		var msg api.StriptMsg
		if err := json.Unmarshal(rawMsg.Content, &msg); err != nil {
			return err
		}

		return c.startScript(msg.Name, msg.Query)
	}
	return api.Error("Invalid msg type")
}

func (c *Conn) RunSync() error {
	// Make authentification message
	authMsg := api.AuthMsg{Login: c.config.Login}
	authMsgBuf, err := json.Marshal(authMsg)
	if err != nil {
		return err
	}

	// Create server
	server := tcpw.ServerDescriptor{
		Raddr:   c.config.URL,
		AuthMsg: authMsgBuf,
	}

	// Connecting
	log.Println("Connecting...")
	connectCtx, connectCancel := context.WithTimeout(c.runCtx, connectTimeout)
	defer connectCancel()
	conn, err := server.Connect(c.runCtx, connectCtx)
	if err != nil {
		return err
	}
	// Starting reader goroutine

	for {
		mt, msg, err := conn.Read(c.runCtx)
		if err != nil {
			return err
		}
		if mt != tcpw.MsgTypeRequest {
			if err := c.conn.Write(c.runCtx, tcpw.MsgTypeError, []byte("Bad request")); err != nil {
				return err
			}
		}
		if err := c.handleRequest(msg); err != nil {
			if err := c.conn.Write(c.runCtx, tcpw.MsgTypeError, []byte(err.Error())); err != nil {
				return err
			}
		}
		if err := c.conn.Write(c.runCtx, tcpw.MsgTypeOk, []byte{}); err != nil {
			return err
		}
	}
}

func (c *Conn) Run() {
	go api.RunAndLog(c.RunSync, "client "+c.config.Login)
}

func (c *Conn) Close() {
	c.stopRun()
}
