package tcp_wrapper

import (
	"context"
	"encoding/binary"
	"net"
	"time"
)

// Some global constants
const (
	readCtxCheckTimeout  = time.Second
	writeCtxCheckTimeout = time.Second
)

//////////////////////////////////// Message types

// Message type constants
type MsgType byte

const (
	MsgTypeUndefined MsgType = iota
	MsgTypeError
	MsgTypeOk
	MsgTypeAuth
	MsgTypeRequest
	MsgTypeResponse
	MsgTypeClose
	MsgTypeDeclined
	MsgTypeAccepted
)

func (mt MsgType) String() string {
	switch mt {
	case MsgTypeUndefined:
		return "Undefined"
	case MsgTypeError:
		return "Error"
	case MsgTypeOk:
		return "Ok"
	case MsgTypeAuth:
		return "Auth"
	case MsgTypeRequest:
		return "Request"
	case MsgTypeResponse:
		return "Response"
	case MsgTypeClose:
		return "Close"
	case MsgTypeDeclined:
		return "Declined"
	case MsgTypeAccepted:
		return "Accepted"
	}
	return "Invalid msg type"
}

//////////////////////////////////// Connection

type Conn struct {
	netConn *net.TCPConn
}

// Validates incomming connection - reads validation message and check it
func (c Conn) validateIn(ctx context.Context) error {
	buf := make([]byte, len(validationMsg))
	if _, err := c.rawRead(ctx, buf); err != nil {
		return err
	}
	if string(buf) != validationMsg { // Checking len is redundant
		return ErrCode(ErrInvalidValidationMsg)
	}
	return nil
}

// Validates outcomming connection - writes validation msg
func (c Conn) validateOut(ctx context.Context) error {
	l, err := c.rawWrite(ctx, []byte(validationMsg))
	if err != nil {
		return err
	}
	if l != len(validationMsg) {
		return ErrCode(ErrInvalidMsgLen) // It would be very strange if it happens...
	}

	return nil
}

// Small wrapper of read that also checks for context
func (c Conn) rawRead(ctx context.Context, buf []byte) (int, error) {
ReadLoop:
	for { // Loop only for context check
		select {
		case <-ctx.Done():
			return 0, ErrCode(ErrContextDone)
		default:
			c.netConn.SetDeadline(time.Now().Add(readCtxCheckTimeout))
			readLen, err := c.netConn.Read(buf)
			if err != nil {
				if err, ok := err.(*net.OpError); ok && err.Timeout() {
					continue ReadLoop
				}
				return readLen, err
			}
			return readLen, nil
		}
	}
}

// Small wrapper of read that also checks for context
func (c Conn) rawWrite(ctx context.Context, buf []byte) (int, error) {
WriteLoop:
	for {
		select {
		case <-ctx.Done():
			return 0, ErrCode(ErrContextDone)
		default:
			c.netConn.SetWriteDeadline(time.Now().Add(writeCtxCheckTimeout))
			l, err := c.netConn.Write(buf)
			if err != nil {
				if err, ok := err.(*net.OpError); ok && err.Timeout() {
					continue WriteLoop
				}
				return l, err
			}
			return l, nil
		}
	}
}

func (c Conn) Read(ctx context.Context) (msgType MsgType, msg []byte, err error) {
	buf := make([]byte, 1024)

	// Read the first part and get the size of the message
	readLen, err := c.rawRead(ctx, buf)
	if err != nil {
		return MsgTypeUndefined, nil, err
	}

	msg = append(msg, buf[5:readLen]...)
	if readLen < 5 {
		return MsgTypeUndefined, msg, ErrCode(ErrInvalidMsgLen)
	}

	msgLen := int(binary.BigEndian.Uint32(buf[0:4])) // Decode the size
	msgType = MsgType(buf[4])

	for readLen < msgLen { // Read the rest
		len := 0

		len, err := c.rawRead(ctx, buf)
		if err != nil {
			return MsgTypeUndefined, msg, err
		}

		msg = append(msg, buf[0:readLen]...)
		readLen += len
	}
	if readLen > msgLen {
		return MsgTypeUndefined, nil, ErrCode(ErrInvalidMsgLen)
	}
	return // All return values are already set
}

func (c Conn) Write(ctx context.Context, msgType MsgType, msg []byte) error {
	buf := make([]byte, 5)
	binary.BigEndian.PutUint32(buf, uint32(len(msg)+5))
	buf[4] = byte(msgType)
	buf = append(buf, msg...) // May be it's too expensive?

	l, err := c.rawWrite(ctx, buf)
	if err != nil {
		return err
	}
	if l != len(buf) { // When can it happen
		return ErrCode(ErrInvalidMsgLen)
	}
	return nil
}

func (c Conn) Close() error {
	return c.netConn.Close()
}
