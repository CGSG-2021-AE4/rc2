package client

import (
	"encoding/json"
	"log"
	"net"
	"os/exec"
	"path"

	cw "github.com/CGSG-2021-AE4/go_utils/conn_wrapper"

	"github.com/CGSG-2021-AE4/rc2/api"
)

type Conn struct {
	configFilename string
	config         *Config

	conn       *cw.Conn
	readerChan chan api.ReadMsg
	doneChan   chan struct{}
}

func NewConnection(configFilename string) (*Conn, error) {
	config, err := LoadConfig(configFilename)
	if err != nil {
		return nil, err
	}

	return &Conn{
		configFilename: configFilename,
		config:         config,

		conn:       nil,
		readerChan: make(chan api.ReadMsg, 3),
		doneChan:   make(chan struct{}),
	}, nil
}

func (c *Conn) register() error {
	// Write registration
	buf, err := json.Marshal(api.RegisterMsg{
		Login: c.config.Login,
	})
	if err != nil {
		return err
	}
	if err := c.conn.Write(cw.MsgTypeRegistration, buf); err != nil {
		return err
	}
	// Wait for response - error/ok
	mt, buf, err := c.conn.Read()
	if err != nil {
		return err
	}
	if mt == cw.MsgTypeClose {
		return api.Error("Close msg: " + string(buf))
	}
	if mt != cw.MsgTypeOk || string(buf) != "Registration complete" {
		return api.Error("Invalid registration responce: " + string(buf))
	}
	return nil
}

func (c *Conn) handleRequest(buf []byte) error {
	var rawMsg api.MainLoopMsg
	if err := json.Unmarshal(buf, &rawMsg); err != nil {
		return err
	}
	if rawMsg.Password != c.config.Password {
		return api.Error("Wrong password")
	}
	switch rawMsg.Type {
	case "script":
		var msg api.StriptMsg
		if err := json.Unmarshal(rawMsg.Content, &msg); err != nil {
			return err
		}
		for i := range len(c.config.Scripts) {
			if c.config.Scripts[i].Name == msg.Name {
				filepath := c.config.Scripts[i].File
				if !path.IsAbs(filepath) {
					filepath = path.Join(path.Dir(c.configFilename), filepath)
				}
				cmd := exec.Command(filepath + " " + msg.Query)
				if err := cmd.Start(); err != nil {
					return err
				}
				return nil
			}
		}
		return api.Error("No script with name: " + msg.Name)
	}
	return api.Error("Message type '" + rawMsg.Type + "' is not supported.")
}

func (c *Conn) Run() error {
	conn, err := net.Dial("tcp", c.config.URL)
	if err != nil {
		return err
	}
	// Send validation msg
	// We are not handling if the validation is wrong because we should not care - it is supposed to work
	if _, err := conn.Write([]byte("TCP forever!!!")); err != nil {
		conn.Close()
		return err
	}

	c.conn = cw.New(conn)
	defer func() {
		c.conn.NetConn.Close()
		c.conn = nil
	}()

	if err := c.register(); err != nil {
		return err
	}
	log.Println("REGISTRATION COMPLETE!!!!")

	// Starting reader goroutine
	go func() (err error) {
		defer func() {
			if err != nil {
				log.Println("End reader cycle with error:", err.Error())
			} else {
				log.Println("End reader cycle")
			}
			close(c.doneChan)
		}()

		for {
			mt, buf, err := c.conn.Read()
			if err != nil {
				return err
			}
			if mt == cw.MsgTypeClose {
				log.Println("CLOSE MSG:", string(buf))
				return nil
			}
			c.readerChan <- api.ReadMsg{Mt: mt, Buf: buf}
		}
	}()

	// Reading cycle
	for {
		select {
		case <-c.doneChan:
			return nil
		case msg := <-c.readerChan:
			if msg.Mt == cw.MsgTypeRequest {
				if err := c.handleRequest(msg.Buf); err != nil {
					if err := c.conn.Write(cw.MsgTypeError, []byte(err.Error())); err != nil {
						return err
					}
				} else {
					if err := c.conn.Write(cw.MsgTypeOk, []byte{}); err != nil {
						return err
					}
				}
			}
		}
	}
}

func (c *Conn) Close() error {
	if c.conn == nil {
		return api.Error("Socket is not connected")
	}
	log.Println("Closing")
	if err := c.conn.Write(cw.MsgTypeClose, []byte("Buy Buy")); err != nil { // Of course it is not thread safe but now I don't care
		return err
	}
	return nil
}
