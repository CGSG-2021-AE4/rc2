package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func NewServer(env EnvVariables) *APIServer {
	apiPtr := &APIServer{
		env:      env,
		DoneChan: make(chan struct{}),
	}
	apiPtr.clientService = NewClientService(apiPtr, env.Host+":"+env.TcpPort)
	apiPtr.msgHandlerService = NewMsgHandlerService(apiPtr)
	apiPtr.statService = NewStatService(apiPtr)
	return apiPtr
}

func (s *APIServer) Run() {
	s.clientService.Serve()

	router := mux.NewRouter()
	router.Handle("/send", s.msgHandlerService)

	s.httpServer = &http.Server{Addr: s.env.Host + ":" + s.env.HttpPort, Handler: router}

	log.Printf("Serving HTTP %s\n", s.env.Host+":"+s.env.HttpPort)
	if err := s.httpServer.ListenAndServe(); err != nil {
		log.Println(err.Error())
	}
	log.Println("END")
}

func (s *APIServer) Close() {
	log.Println("BBBB")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	close(s.DoneChan)
	s.httpServer.Shutdown(ctx)
	if err := s.clientService.Close(); err != nil {
		log.Println("Client service closed with error:", err)
	}
}
