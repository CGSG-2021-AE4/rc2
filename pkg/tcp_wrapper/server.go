package tcp_wrapper

import (
	"context"
	"net"
	"time"
)

const (
	acceptTimeout = 2 * time.Second
	listenTimeout = 2 * time.Second

	validationMsg = "TCP"
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

ContextCheckLoop: // Is needed only!!! to check context
	for {
		select {
		case <-ctx.Done():
			return nil, ErrCode(ErrContextDone)
		default:
			//// Accepting

			ts.listener.SetDeadline(time.Now().Add(acceptTimeout))
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
			if err := conn.validateIn(handshakeCtx); err != nil {
				return nil, err
			}
			// Read authentification message
			authMsgType, authMsg, err := conn.Read(handshakeCtx)
			if err != nil {
				return nil, err
			}
			if authMsgType != MsgTypeAuth {
				return nil, ErrCode(ErrInvalidAuthMsgType)
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
