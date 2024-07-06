package api

import "encoding/json"

// Common functionality for both server and client

/////////////// Messages' structs

// register
type RegisterMsg struct {
	Login string `json:"login"`
}

// Msg to send to client
type SendRequestMsg struct {
	Login string          `json:"login"`
	Msg   json.RawMessage `json:"msg"`
}

// Messages' structs
type ReadMsg struct {
	Mt  byte
	Buf []byte
}
