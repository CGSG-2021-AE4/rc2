package client

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/CGSG-2021-AE4/rc2/api"
	tcpw "github.com/CGSG-2021-AE4/rc2/pkg/tcp_wrapper"
)

const (
	reconnectTimeout = 3 * time.Second
)

type RequestService struct {
	config *Config
	mutex  sync.Mutex

	conn *Conn

	runCtx  context.Context
	stopRun context.CancelFunc
}

func NewReqService(config *Config) *RequestService {
	ctx, cancel := context.WithCancel(context.Background())
	return &RequestService{
		config: config,

		runCtx:  ctx,
		stopRun: cancel,
	}
}

func (rs *RequestService) Conn() *Conn {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	return rs.conn
}

func (rs *RequestService) SetConn(c *Conn) {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	rs.conn = c
}

func (rs *RequestService) RunSync() error {
	var reconnectWait <-chan time.Time = nil

ConnectLoop:
	for {
		select {
		case <-rs.runCtx.Done():
			return tcpw.ErrCode(tcpw.ErrContextDone)
		default:
			if reconnectWait != nil {
				<-reconnectWait
				log.Println("Reconnecting...")
			}
			conn, err := NewConn(rs.config)
			if err != nil {
				log.Println("Faild to create connection:", err.Error())
				continue ConnectLoop
			}
			rs.SetConn(conn)
			api.RunAndLog(conn.RunSync, "client connection")
			rs.SetConn(nil)
			reconnectWait = time.After(reconnectTimeout)
		}
	}
}

func (rs *RequestService) Run() {
	api.RunAndLog(rs.RunSync, "request service")
}

func (rs *RequestService) Close() {
	// Close connection
	c := rs.Conn()
	if c != nil {
		c.Close()
	}
	// Stop cycle
	rs.stopRun()
}
