package tcp_wrapper

import (
	"context"
	"encoding/binary"
	"net"
	"time"
)

// TCP functionality
// Temporary it is here

// Some global constants
const (
	acceptTimeout = 4 * time.Second
	listenTimeout = 2 * time.Second
	readTimeout   = time.Second
	writeTimeout  = time.Second

	validationMsg = "TCP"
)

// Message type constants

const (
	MsgTypeUndefined byte = iota // if error while reading
	MsgTypeError
	MsgTypeOk
	MsgTypeAuth
	MsgTypeRequest
	MsgTypeResponse
	MsgTypeClose
	MsgTypeDeclined
	MsgTypeAccepted
)

////////////////////////////////// Server

type TcpServer struct {
	listener *net.TCPListener
}

func NewTcpServer(listenAddr string) (*TcpServer, error) {
	listenner, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}
	return &TcpServer{
		listener: listenner.(*net.TCPListener),
	}, nil
}

// It must be executed to deinit not to stop!
func (ts *TcpServer) Close() error {
	return ts.listener.Close()
}

func (ts *TcpServer) Listen(ctx context.Context) (*ConnAcceptor, error) {
	// It is a bit shity because I block listen thread to make a handshake

	ts.listener.SetDeadline(time.Now().Add(acceptTimeout))

ContextCheckLoop: // Is needed only!!! to check context
	for {
		select {
		case <-ctx.Done():
			return nil, ErrContextDone
		default:
			//// Accepting

			netConn, err := ts.listener.AcceptTCP()
			if err != nil {
				if err, ok := err.(*net.OpError); ok && err.Timeout() {
					continue ContextCheckLoop
				}
				return nil, err
			}
			conn := Conn{netConn: netConn}

			handshakeCtx, cancel := context.WithTimeout(ctx, acceptTimeout) // Create listen context with timeout
			defer cancel()

			//// Handshake

			// Validate connection
			if err := conn.validate(handshakeCtx); err != nil {
				return nil, err
			}
			// Read authentification message
			authMsgType, authMsg, err := conn.Read(handshakeCtx)
			if err != nil {
				return nil, err
			}
			if authMsgType != MsgTypeAuth {
				return nil, ErrInvalidAuthMsgType
			}
			return &ConnAcceptor{conn: &conn, AuthMsg: authMsg}, nil
		}
	}
}

//////////////////////////////////// Acceptor

type ConnAcceptor struct {
	AuthMsg []byte
	conn    *Conn // May be later there will be more data than just NetConn
}

func (ca *ConnAcceptor) Accept(ctx context.Context) (*Conn, error) {
	if err := ca.conn.Write(ctx, MsgTypeAccepted, []byte{}); err != nil { // Send to client data that it is accepted
		return nil, err
	}
	return ca.conn, nil
}

func (ca *ConnAcceptor) Decline(ctx context.Context, declineMsg []byte) error {
	if err := ca.conn.Write(ctx, MsgTypeDeclined, declineMsg); err != nil { // Send to client data that it is declined
		return err
	}
	return ca.conn.Close()
}

//////////////////////////////////// Connection

type Conn struct {
	netConn *net.TCPConn
}

// Reads validation message and compare it with the right one
func (c Conn) validate(ctx context.Context) error {
	buf := make([]byte, len(validationMsg))
	select {
	case <-ctx.Done():
		return ErrContextDone
	default:
		c.netConn.SetDeadline(time.Now().Add(readTimeout))
		_, err := c.netConn.Read(buf)
		if err != nil {
			return err
		}
	}
	if string(buf) != validationMsg { // Checking len is redundant
		return ErrInvalidValidationMsg
	}
	return nil
}

func (c Conn) Read(ctx context.Context) (msgType byte, msg []byte, err error) {
	buf := make([]byte, 1024)

	// Read the first part and get the size of the message
	readLen := 0

FirstReadLoop:
	for { // Loop only for context check
		select {
		case <-ctx.Done():
			return MsgTypeUndefined, nil, ErrContextDone
		default:
			c.netConn.SetDeadline(time.Now().Add(readTimeout))
			readLen, err = c.netConn.Read(buf)
			if err != nil {
				if err, ok := err.(*net.OpError); ok && err.Timeout() {
					continue FirstReadLoop
				}
				return MsgTypeUndefined, nil, err
			}
			break FirstReadLoop
		}
	}

	msg = append(msg, buf[5:readLen]...)
	if readLen < 5 {
		return MsgTypeUndefined, msg, ErrInvalidMsgLen
	}

	msgLen := int(binary.BigEndian.Uint32(buf[0:4])) // Decode the size
	msgType = buf[4]

	for readLen < msgLen { // Read the rest
		len := 0

		// Breakable read...
	SecondReadLoop:
		for { // Loop only for context check
			select {
			case <-ctx.Done():
				return MsgTypeUndefined, msg, ErrContextDone
			default:
				c.netConn.SetDeadline(time.Now().Add(readTimeout))
				len, err = c.netConn.Read(buf)
				if err != nil {
					if err, ok := err.(*net.OpError); ok && err.Timeout() {
						continue SecondReadLoop
					}
					return MsgTypeUndefined, msg, err
				}
				break SecondReadLoop
			}
		}
		// End of breakable read

		msg = append(msg, buf[0:readLen]...)
		readLen += len
	}
	if readLen > msgLen {
		return MsgTypeUndefined, nil, ErrInvalidMsgLen
	}
	return // All return values are already set
}

func (c Conn) Write(ctx context.Context, msgType byte, msg []byte) error {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf, uint32(len(msg)+5))
	buf[4] = msgType
	buf = append(buf, msg...) // May be it's too expensive?

	// Breakable write
WriteLoop:
	for {
		select {
		case <-ctx.Done():
			return ErrContextDone
		default:
			c.netConn.SetWriteDeadline(time.Now().Add(writeTimeout))
			l, err := c.netConn.Write(buf)
			if err != nil {
				if err, ok := err.(*net.OpError); ok && err.Timeout() {
					continue WriteLoop
				}
				return err
			}
			if l != len(buf) { // When can it happen
				return ErrInvalidMsgLen
			}
			return nil
		}
	}
}

func (c Conn) Close() error {
	return c.netConn.Close()
}
