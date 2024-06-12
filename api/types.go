package api

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

// Main API server
type APIServer struct {
	listenAddr        string
	clientService     *ClientService
	msgHandlerService *MsgHandlerService
}

// Serves clients' websocket connections
type ClientService struct {
	conns map[string]*ClientConn // SHIT not thread safe
}

// Client connection
type ClientConn struct {
	cs         *ClientService
	conn       *websocket.Conn
	login      string
	readerChan chan []byte
}

// Message handler service
type MsgHandlerService struct {
	s *APIServer
}

// My error implementation
type rcError struct {
	err string
}

/////////////// Messages' structs

// register
type registerMsg struct {
	Login string `json:"login"`
}

// Msg to send to client
type sendRequestMsg struct {
	Login string          `json:"login"`
	Msg   json.RawMessage `json:"msg"`
}
