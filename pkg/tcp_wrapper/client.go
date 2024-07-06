package tcp_wrapper

import (
	"context"
	"net"
	"time"
)

const (
	dialTimeout     = 2 * time.Second
	handshakeTimout = 2 * time.Second
)

//////////////////////////////////// ServerDescriptor

// Just describes connection params
type ServerDescriptor struct {
	Raddr   string // Remote addr
	AuthMsg []byte // Message for authrization
}

func (c *ServerDescriptor) Connect(dialCtx context.Context, connectCtx context.Context) (outConn *Conn, outErr error) {
	// Connection itself
	dialer := net.Dialer{Timeout: dialTimeout}
	netConn, err := dialer.DialContext(dialCtx, "tcp", c.Raddr)
	if err != nil {
		return nil, err
	}
	conn := Conn{netConn: netConn.(*net.TCPConn)}

	handshakeCtx, cancel := context.WithTimeout(connectCtx, handshakeTimout)
	defer func() {
		if tcpErr, ok := outErr.(TcpError); ok && tcpErr.Code == ErrContextDone { // Check if it was timeout, not the context cancelation
			select {
			case <-connectCtx.Done():
				break
			case <-handshakeCtx.Done():
				outErr = ErrCode(ErrTimeout)
				break
			}
		}

		cancel()
	}()

	// Sending validation msg
	if err := conn.validateOut(handshakeCtx); err != nil {
		return nil, err
	}
	// Sending auth msg
	if err := conn.Write(handshakeCtx, MsgTypeAuth, c.AuthMsg); err != nil {
		return nil, err
	}
	// Waiting for the answer
	msgType, msg, err := conn.Read(handshakeCtx)
	if err != nil {
		return nil, err
	}
	switch msgType {
	case MsgTypeAccepted:
		return &Conn{netConn: netConn.(*net.TCPConn)}, nil
	case MsgTypeDeclined:
		return nil, ErrDescr(ErrConnectionDeclined, string(msg)) // For this only line I changed all errors creation...
	default:
		return nil, ErrCode(ErrInvalidMsgLen)
	}
}
