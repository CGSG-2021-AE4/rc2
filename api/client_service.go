package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{}

// Constructor
func NewClientService() *ClientService {
	return &ClientService{
		conns: map[string]*ClientConn{},
	}
}

func (cs *ClientService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	client := NewClient(cs, c)
	go func() {
		client.Run()
		log.Println("Connnection ", c.RemoteAddr(), " closed.")
	}()
}

func NewClient(cs *ClientService, c *websocket.Conn) *ClientConn {
	return &ClientConn{
		cs:         cs,
		conn:       c,
		login:      "",
		readerChan: make(chan []byte, 3),
	}
}

func (c *ClientConn) WriteError(errMsg string) {
	log.Println("WRITE ERROR: ", errMsg)
	msg, err := WriteError(errMsg)
	if err != nil {
		log.Println(err)
		return
	}
	if err := c.conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
		log.Println(err)
	}
}

func (c *ClientConn) register() (outErr error, notifyClient bool) {
	// Registration
	// Wait for registration request
	wsmt, buf, err := c.conn.ReadMessage()
	if err != nil {
		return err, false
	}
	// Check that it is registration
	if wsmt != websocket.BinaryMessage {
		return NewError("Invalid registration message type"), true
	}
	mt, rawMsg, err := ReadMsg[json.RawMessage](buf)
	if err != nil {
		return NewError("Invalid json"), true
	}
	if mt != "registration" {
		return NewError("Invalid registration message type"), true
	}
	var msg registerMsg
	if err := json.Unmarshal(rawMsg, &msg); err != nil {
		return err, true
	}
	// Check if such a login is already registered
	if c.cs.conns[msg.Login] != nil {
		return NewError("Double registration"), true
	}
	// Register login
	c.login = msg.Login
	c.cs.conns[c.login] = c
	// Notify that is fine
	completeMsg, err := WriteMsg("msg", "Registration complete")
	if err != nil {
		return err, false
	}
	if err := c.conn.WriteMessage(websocket.BinaryMessage, completeMsg); err != nil {
		return err, false
	}
	log.Println("registraionComplete")
	return nil, false
}

func (c *ClientConn) Run() (outErr error, notifyClient bool) {
	defer func() {
		if outErr != nil {
			log.Println("Connnection ", c.conn.RemoteAddr(), " closed with error: ", outErr.Error())
			if notifyClient {
				c.conn.WriteMessage(websocket.CloseMessage, []byte(outErr.Error()))
			}
		} else {
			c.conn.WriteMessage(websocket.CloseMessage, []byte("Bye"))
		}
		if c.login != "" && c.cs.conns[c.login] != nil {
			delete(c.cs.conns, c.login)
			log.Println("Unregistered:", c.login)
		}
		close(c.readerChan)
		c.conn.Close()
	}()

	if err, doNotify := c.register(); err != nil {
		return err, doNotify
	}

	for {
		wsmt, buf, err := c.conn.ReadMessage()
		if err != nil {
			return err, false
		}
		if wsmt == websocket.CloseMessage {
			break
		}
		c.readerChan <- buf
	}
	return nil, false
}

func (c *ClientConn) WriteMsg(buf []byte) ([]byte, error) {
	if err := c.conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
		return nil, err
	}
	ans := <-c.readerChan
	return ans, nil
}
