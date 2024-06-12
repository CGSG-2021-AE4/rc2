package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func NewServer(listenAddr string) *APIServer {
	apiPtr := &APIServer{
		listenAddr:    listenAddr,
		clientService: NewClientService(),
	}
	apiPtr.msgHandlerService = NewMsgHandlerService(apiPtr)
	return apiPtr
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.Handle("/client_service", s.clientService)
	router.Handle("/send", s.msgHandlerService)

	fmt.Printf("Serving %s\n", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}
