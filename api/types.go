package api

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"
	"time"

	cw "github.com/CGSG-2021-AE4/go_utils/conn_wrapper"
)

// All environment variables that can be set with start arguments
type EnvVariables struct {
	Host           string
	HttpPort       string
	TcpPort        string
	EnableStateLog bool
	StatisticsFile string
	StatTimeout    time.Duration // In MS
}

// Main API server
type APIServer struct {
	env               EnvVariables
	httpServer        *http.Server
	clientService     *ClientService
	msgHandlerService *MsgHandlerService
	statService       *StatService
	DoneChan          chan struct{} // Chanel that make all other threads like stat service stop
}

// Serves clients' websocket connections
type ClientService struct {
	server     *APIServer
	listenAddr string
	listener   net.Listener
	connMutex  sync.Mutex
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

/////////////// Statistics

type statPerUser struct {
	Login    string        `json:"login"`
	LastAddr string        `json:"lastIp"`
	WorkDur  time.Duration `json:"workDuration"`
	LastSeen time.Time     `json:"lastSeen"`
}

type stat struct {
	Host      string        `json:"host"`
	WorkDur   time.Duration `json:"workDuration"`
	Started   time.Time     `json:"started"`
	WriteTime time.Time     `json:"writeTime"`
	Users     []statPerUser `json:"users"`
}

type StatService struct {
	server         *APIServer
	currentStat    stat
	startTime      time.Time
	Mutex          sync.Mutex // Locks all local data
	connectedUsers map[string]connectedUserStat
}

type connectedUserStat struct {
	connStartTime time.Time
	addr          string
}
