package api

import (
	"encoding/json"
	"log"
	"net"

	cw "github.com/CGSG-2021-AE4/go_utils/conn_wrapper"
)

// Constructor
func NewClientService(listenAddr string) *ClientService {
	return &ClientService{
		listenAddr: listenAddr,
		conns:      map[string]*ClientConn{},
	}
}

func (cs *ClientService) Serve() error {
	log.Println("TCP server: ", cs.listenAddr)
	listener, err := net.Listen("tcp", cs.listenAddr)
	if err != nil {
		return err
	}

	// Start accept cycle
	go func() (err error) {
		defer func() {
			if err != nil {
				log.Println("End accept cycle with error:", err.Error())
			} else {
				log.Println("End accept cycle")
			}
		}()

		for {
			c, err := listener.Accept()
			if err != nil {
				return err
			}
			log.Println("New conn: ", c.RemoteAddr().Network(), c.RemoteAddr())
			go func(c net.Conn) {
				client := NewClient(cs, cw.NewConn(c))
				client.Run()
				log.Println("Connnection ", c.RemoteAddr().Network(), c.RemoteAddr(), " closed.")
			}(c)
		}
	}()
	return nil
}

func NewClient(cs *ClientService, c *cw.ConnWrapper) *ClientConn {
	return &ClientConn{
		cs:         cs,
		conn:       c,
		login:      "",
		readerChan: make(chan []byte, 3),
	}
}

func (c *ClientConn) WriteError(msg string) {
	log.Println("WRITE ERROR: ", msg)
	if err := c.conn.Write(cw.MsgTypeError, []byte(msg)); err != nil {
		log.Println(err)
	}
}

func (c *ClientConn) register() (outErr error, notifyClient bool) {
	// Registration
	// Wait for registration request
	mt, buf, err := c.conn.Read()
	if err != nil {
		return err, true
	}
	if mt != cw.MsgTypeRegistration {
		return rcError("Invalid message type"), true
	}
	var msg registerMsg
	if err := json.Unmarshal(buf, &msg); err != nil {
		return err, true
	}
	// Check if such a login is already registered
	if c.cs.conns[msg.Login] != nil {
		return rcError("Double registration"), true
	}
	// Register login
	c.login = msg.Login
	c.cs.conns[c.login] = c
	// Notify that is fine
	if err := c.conn.Write(cw.MsgTypeOk, []byte("Registration complete")); err != nil {
		return err, false
	}
	log.Println("Registraion complete")
	return nil, false
}

func (c *ClientConn) Run() (outErr error, notifyClient bool) {
	defer func() {
		if outErr != nil {
			log.Println("Connnection ", c.conn.Conn.RemoteAddr(), " closed with error: ", outErr.Error())
			if notifyClient {
				c.conn.Write(cw.MsgTypeClose, []byte(outErr.Error()))
			}
		} else {
			c.conn.Write(cw.MsgTypeClose, []byte("Bye"))
		}
		if c.login != "" && c.cs.conns[c.login] != nil {
			delete(c.cs.conns, c.login)
			log.Println("Unregistered:", c.login)
		}
		close(c.readerChan)
		c.conn.Conn.Close()
	}()

	if err, doNotify := c.register(); err != nil {
		return err, doNotify
	}

	for {
		mt, buf, err := c.conn.Read()
		if err != nil {
			return err, false
		}
		log.Println("Got msg", mt, buf)
		if mt == cw.MsgTypeClose {
			break
		}
		c.readerChan <- buf
	}
	return nil, false
}

func (c *ClientConn) WriteMsg(buf []byte) ([]byte, error) {
	if err := c.conn.Write(cw.MsgTypeRequest, buf); err != nil {
		return nil, err
	}
	ans := <-c.readerChan
	return ans, nil
}
