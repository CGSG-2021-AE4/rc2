package api

import (
	"encoding/json"
	"log"
	"net"

	cw "github.com/CGSG-2021-AE4/go_utils/conn_wrapper"
)

// Constructor
func NewClientService(server *APIServer, listenAddr string) *ClientService {
	return &ClientService{
		server:     server,
		listenAddr: listenAddr,
		conns:      make(map[string]*ClientConn),
	}
}

func (cs *ClientService) Serve() error {
	log.Println("TCP server: ", cs.listenAddr)
	listener, err := net.Listen("tcp", cs.listenAddr)
	if err != nil {
		return err
	}
	cs.listener = listener

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
			c, err := cs.listener.Accept()
			if err != nil {
				return err
			}
			// !!! Ensure that it's right tcp connection
			// Validation msg: "TCP forever!!!"
			validationMsg := "TCP forever!!!"
			buf := make([]byte, len(validationMsg))
			if _, err := c.Read(buf); err != nil || string(buf) != validationMsg {
				log.Println("Invalid validation msg:", string(buf))
				c.Close()
				continue
			}
			log.Println("New conn: ", c.RemoteAddr().Network(), c.RemoteAddr())
			go func(c net.Conn) {
				client := NewClient(cs, cw.New(c))
				client.Run()
				log.Println("Connnection ", c.RemoteAddr().Network(), c.RemoteAddr(), " closed.")
			}(c)
		}
	}()
	return nil
}

func (cs *ClientService) GetCurClients() []*ClientConn {
	cs.connMutex.Lock()
	defer cs.connMutex.Unlock()

	conns := make([]*ClientConn, len(cs.conns))
	i := 0
	for _, c := range cs.conns {
		conns[i] = c
		i++
	}
	return conns
}

func (cs *ClientService) Close() error {
	if cs.listener != nil {
		log.Println("Close TCP listener")
		return cs.listener.Close()
	}
	log.Println("Close TCP clistener is NIL")
	return nil
}

/////////////// Client connection

func NewClient(cs *ClientService, c *cw.Conn) *ClientConn {
	return &ClientConn{
		cs:         cs,
		conn:       c,
		isOpen:     false,
		login:      "",
		readerChan: make(chan readMsg, 5),
		doneChan:   make(chan struct{}),
	}
}

func (c *ClientConn) WriteError(msg string) {
	log.Println("WRITE ERROR: ", msg)
	if err := c.conn.Write(cw.MsgTypeError, []byte(msg)); err != nil {
		log.Println(err)
	}
}

func (c *ClientConn) register() error {
	// Registration
	// Wait for registration request
	var msg readMsg
	select {
	case <-c.doneChan:
		return rcError("Unexpected done")
	case msg = <-c.readerChan:
		break
	}
	if msg.mt != cw.MsgTypeRegistration {
		return rcError("Invalid message type")
	}
	var regMsg registerMsg
	if err := json.Unmarshal(msg.buf, &regMsg); err != nil {
		return err
	}
	// Check if such a login is already registered
	c.cs.connMutex.Lock()
	defer c.cs.connMutex.Unlock() // But I spend time on writing...

	if c.cs.conns[regMsg.Login] != nil {
		return rcError("Double registration")
	}
	// Register login
	c.login = regMsg.Login
	c.cs.conns[c.login] = c

	// Notify that is fine
	if err := c.conn.Write(cw.MsgTypeOk, []byte("Registration complete")); err != nil {
		return err
	}
	log.Println("Registraion complete")
	return nil
}

func (c *ClientConn) readCycle() (err error) {
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
			log.Println("INVALID MSG: ", mt, string(buf), err.Error())
			return err
		}
		if mt == cw.MsgTypeClose {
			log.Println("CLOSE MSG:", string(buf))
			return nil
		}
		c.readerChan <- readMsg{mt, buf}
	}
}

func (c *ClientConn) Run() (err error) {
	defer func() {
		defer func() { // omg defer in defer... but I have to close channels
			close(c.readerChan)
			c.cs.server.statService.OnDisconnect(c)
		}()
		// Run will log error of close here and return error of closing if it occurred
		if err != nil { // If there is error
			log.Println("Connnection ", c.conn.NetConn.RemoteAddr(), " closed with error: ", err.Error())
			if c.isOpen { // If connection is still open server will send error msg
				if err = c.conn.Write(cw.MsgTypeError, []byte(err.Error())); err != nil {
					return
				}
			}
		}
		c.cs.connMutex.Lock()
		defer c.cs.connMutex.Unlock() // But I spend time on writing...

		if c.login != "" && c.cs.conns[c.login] != nil {
			delete(c.cs.conns, c.login)
			log.Println("Unregistered:", c.login)
		}
		if c.isOpen { // Is still open => close
			if err = c.conn.Write(cw.MsgTypeClose, []byte("Bye")); err != nil {
				return
			}
			err = c.conn.NetConn.Close()
		}
	}()
	c.cs.server.statService.OnConnect(c)

	// Starting reader goroutine
	go c.readCycle()

	if err := c.register(); err != nil {
		return err
	}

	for range c.doneChan { // It seams I just have to wait untill I have to close connection...
		return nil
	}
	return rcError("Reach return after infinit cycle.")
}

func (c *ClientConn) WriteMsg(buf []byte) (readMsg, error) {
	if err := c.conn.Write(cw.MsgTypeRequest, buf); err != nil {
		return readMsg{cw.MsgTypeUndefined, nil}, err
	}
	return <-c.readerChan, nil
}
