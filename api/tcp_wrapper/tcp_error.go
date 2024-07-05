package tcp_wrapper

type tcpError byte

func (err tcpError) Error() string {
	switch err {
	case ErrContextDone:
		return "Context done"
	case ErrAcceptTimeout:
		return "Accept timeout"
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

// Some error contants

const ( // TODO change to int codes
	ErrContextDone = tcpError(iota)
	ErrAcceptTimeout
	ErrHandshakeFailed
	ErrInvalidValidationMsg
	ErrInvalidAuthMsgType
	ErrInvalidMsgLen
)
