package api

import (
	"encoding/json"
	"log"
	"net/http"
)

type sendRequest struct {
	Login string          `json:"login"`
	Msg   json.RawMessage `json:"msg"`
}

type MsgHandler struct {
	s *APIServer
}

func NewMsgHandler(s *APIServer) *MsgHandler {
	return &MsgHandler{
		s: s,
	}
}

func (mh *MsgHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body sendRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	c := mh.s.clientService.conns[body.Login]
	if c == nil {
		http.Error(w, "Client is not connected", http.StatusBadRequest)
		return
	}

	response, err := c.WriteMsg(body.Msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(response); err != nil {
		log.Println(err)
	}
}
