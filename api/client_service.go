package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{}

// Serves clients' websocket connections
type ClientService struct {
	conns map[string]*ClientConn // SHIT not thread safe
}

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
	go client.Run()
}

// Client connection
type ClientConn struct {
	cs    *ClientService
	conn  *websocket.Conn
	login string
}

func NewClient(cs *ClientService, c *websocket.Conn) *ClientConn {
	return &ClientConn{
		cs:    cs,
		conn:  c,
		login: "",
	}
}

// Messages' structs
type registerMsg struct { // register
	Login string `json:"login"`
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

func (c *ClientConn) Run() {
	defer c.conn.Close()
	c.conn.SetCloseHandler(func(code int, text string) error {
		log.Println("Close connection.")
		if c.login != "" && c.cs.conns[c.login] != nil {
			delete(c.cs.conns, c.login)
		}
		return nil
	})

	// Registration
	registraionComplete := false
	for !registraionComplete {
		// Wait for registration request
		wsmt, buf, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("READ ERROR: ", err)
			return
		}
		// Check that it is registration
		if wsmt != websocket.BinaryMessage {
			if wsmt == websocket.CloseMessage {
				return
			}
			c.WriteError("Invalid msg type: wait for registration")
			continue
		}
		mt, rawMsg, err := ReadMsg[json.RawMessage](buf)
		if err != nil {
			c.WriteError("Invalid json")
			continue
		}
		if mt != "registration" {
			c.WriteError("Invalid msg type: wait for registration")
			continue
		}
		var msg registerMsg
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			log.Println(err)
			return
		}
		// Check if such a login is already registered
		if c.cs.conns[msg.Login] != nil {
			c.WriteError("Double registration")
			continue
		}
		// Register login
		c.cs.conns[msg.Login] = c
		registraionComplete = true
		// Notify that is fine
		completeMsg, err := WriteMsg("msg", "Registration complete")
		if err != nil {
			log.Println(err)
			return
		}
		if err := c.conn.WriteMessage(websocket.BinaryMessage, completeMsg); err != nil {
			log.Println(err)
			return
		}
	}
	log.Println(registraionComplete)

	for {
		wsmt, buf, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("READ ERROR: ", err)
			if websocket.IsCloseError(err) {
				return
			}
			continue
		}
		log.Println("GO MSG: ", websocket.FormatMessageType(wsmt), buf)
	}
}
