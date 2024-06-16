package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func NewServer(host, httpPort, tcpPort string) *APIServer {
	apiPtr := &APIServer{
		host:          host,
		httpPort:      httpPort,
		tcpPort:       tcpPort,
		clientService: NewClientService(host + ":" + tcpPort),
	}
	apiPtr.msgHandlerService = NewMsgHandlerService(apiPtr)
	return apiPtr
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	s.clientService.Serve()
	router.Handle("/send", s.msgHandlerService)

	log.Printf("Serving HTTP %s\n", s.host+":"+s.httpPort)
	if err := http.ListenAndServe(s.host+":"+s.httpPort, router); err != nil {
		log.Println(err.Error())
	}
	log.Println("END")
}
