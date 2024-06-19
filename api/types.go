package api

import (
	"encoding/json"

	cw "github.com/CGSG-2021-AE4/go_utils/conn_wrapper"
)

// Main API server
type APIServer struct {
	host              string
	httpPort          string
	tcpPort           string
	clientService     *ClientService
	msgHandlerService *MsgHandlerService
}

// Serves clients' websocket connections
type ClientService struct {
	listenAddr string
	conns      map[string]*ClientConn // SHIT not thread safe
}

// Messages' structs
type readMsg struct {
	mt  byte
	buf []byte
}

// Client connection
type ClientConn struct {
	cs         *ClientService
	conn       *cw.Conn
	isOpen     bool
	login      string
	readerChan chan readMsg
	doneChan   chan struct{}
}

// Message handler service
type MsgHandlerService struct {
	s *APIServer
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
