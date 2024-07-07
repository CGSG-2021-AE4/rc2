package api

import "encoding/json"

// Common functionality for both server and client

/////////////// Messages' structs

// Auth
type AuthMsg struct {
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

type StriptMsg struct { // stript
	Name  string `json:"name"`
	Query string `json:"query"`
}

type MainLoopMsg struct {
	Password string          `json:"password"`
	Type     string          `json:"type"`
	Content  json.RawMessage `json:"content"`
}
