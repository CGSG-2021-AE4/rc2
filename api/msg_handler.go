package api

import (
	"encoding/json"
	"log"
	"net/http"

	cw "github.com/CGSG-2021-AE4/go_utils/conn_wrapper"
)

func NewMsgHandlerService(s *APIServer) *MsgHandlerService {
	return &MsgHandlerService{
		s: s,
	}
}

func (mh *MsgHandlerService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var body sendRequestMsg
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

	if _, err := w.Write([]byte("Response type: " + cw.FormatError(response.mt) + ", Msg: " + string(response.buf))); err != nil { // TODO format msg type as well
		log.Println(err)
	}
}
