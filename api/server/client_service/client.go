package client_service

import (
	"encoding/json"
	"log"

	cw "github.com/CGSG-2021-AE4/go_utils/conn_wrapper"

	"github.com/CGSG-2021-AE4/rc2/api"
)

// Client connection
type Conn struct {
	cs         *Service
	Conn       *cw.Conn
	IsOpen     bool
	Login      string
	readerChan chan api.ReadMsg
	doneChan   chan struct{}
}

func NewClient(cs *Service, c *cw.Conn) *Conn {
	return &Conn{
		cs:         cs,
		Conn:       c,
		IsOpen:     false,
		Login:      "",
		readerChan: make(chan api.ReadMsg, 5),
		doneChan:   make(chan struct{}),
	}
}

func (c *Conn) WriteError(msg string) {
	log.Println("WRITE ERROR: ", msg)
	if err := c.Conn.Write(cw.MsgTypeError, []byte(msg)); err != nil {
		log.Println(err)
	}
}

func (c *Conn) register() error {
	// Registration
	// Wait for registration request
	var msg api.ReadMsg
	select {
	case <-c.doneChan:
		return api.Error("Unexpected done")
	case msg = <-c.readerChan:
		break
	}
	if msg.Mt != cw.MsgTypeRegistration {
		return api.Error("Invalid message type")
	}
	var regMsg api.RegisterMsg
	if err := json.Unmarshal(msg.Buf, &regMsg); err != nil {
		return err
	}
	// Check if such a login is already registered
	c.cs.connMutex.Lock()
	defer c.cs.connMutex.Unlock() // But I spend time on writing...

	if c.cs.Conns[regMsg.Login] != nil {
		return api.Error("Double registration")
	}
	// Register login
	c.Login = regMsg.Login
	c.cs.Conns[c.Login] = c

	// Notify that is fine
	if err := c.Conn.Write(cw.MsgTypeOk, []byte("Registration complete")); err != nil {
		return err
	}
	log.Println("Registraion complete")
	return nil
}

func (c *Conn) readCycle() (err error) {
	defer func() {
		if err != nil {
			log.Println("End reader cycle with error:", err.Error())
		} else {
			log.Println("End reader cycle")
		}
		close(c.doneChan)
	}()

	for {
		mt, buf, err := c.Conn.Read()
		if err != nil {
			log.Println("INVALID MSG: ", mt, string(buf), err.Error())
			return err
		}
		if mt == cw.MsgTypeClose {
			log.Println("CLOSE MSG:", string(buf))
			return nil
		}
		c.readerChan <- api.ReadMsg{Mt: mt, Buf: buf}
	}
}

func (c *Conn) Run() (err error) {
	defer func() {
		defer func() { // omg defer in defer... but I have to close channels
			close(c.readerChan)
			// c.cs.server.StatService.OnDisconnect(c) TODO
		}()
		// Run will log error of close here and return error of closing if it occurred
		if err != nil { // If there is error
			log.Println("Connnection ", c.Conn.NetConn.RemoteAddr(), " closed with error: ", err.Error())
			if c.IsOpen { // If connection is still open server will send error msg
				if err = c.Conn.Write(cw.MsgTypeError, []byte(err.Error())); err != nil {
					return
				}
			}
		}
		c.cs.connMutex.Lock()
		defer c.cs.connMutex.Unlock() // But I spend time on writing...

		if c.Login != "" && c.cs.Conns[c.Login] != nil {
			delete(c.cs.Conns, c.Login)
			log.Println("Unregistered:", c.Login)
		}
		if c.IsOpen { // Is still open => close
			if err = c.Conn.Write(cw.MsgTypeClose, []byte("Bye")); err != nil {
				return
			}
			err = c.Conn.NetConn.Close()
		}
	}()
	// c.cs.server.StatService.OnConnect(c) TODO

	// Starting reader goroutine
	go c.readCycle()

	if err := c.register(); err != nil {
		return err
	}

	for range c.doneChan { // It seams I just have to wait untill I have to close connection...
		return nil
	}
	return api.Error("Reach return after infinit cycle.")
}

func (c *Conn) WriteMsg(buf []byte) (api.ReadMsg, error) {
	if err := c.Conn.Write(cw.MsgTypeRequest, buf); err != nil {
		return api.ReadMsg{Mt: cw.MsgTypeUndefined, Buf: nil}, err
	}
	return <-c.readerChan, nil
}
