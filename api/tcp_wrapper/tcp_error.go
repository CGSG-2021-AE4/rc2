package tcp_wrapper

// Error codes
const ( // TODO change to int codes
	ErrContextDone byte = iota
	ErrTimeout
	ErrConnectionDeclined
	ErrHandshakeFailed
	ErrInvalidValidationMsg
	ErrInvalidAuthMsgType
	ErrInvalidMsgLen
)

// Error type with description
type TcpError struct {
	Code  byte
	Descr string
}

func (err TcpError) Error() string {
	switch err.Code {
	case ErrContextDone:
		return "Context done"
	case ErrTimeout:
		return "Timeout"
	case ErrHandshakeFailed:
		return "Handshake failed"
	case ErrInvalidValidationMsg:
		return "Invalid validation message"
	case ErrInvalidAuthMsgType:
		return "Invalid authentification message type"
	case ErrInvalidMsgLen:
		return "Invalid message len"
	}
	return "Undefined error code"
}

// New error functions that does not set description
func ErrCode(code byte) TcpError {
	return TcpError{Code: code}
}

// Net error function that sets both code and description
func ErrDescr(code byte, descr string) TcpError {
	return TcpError{Code: code, Descr: descr}
}
