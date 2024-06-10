package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr    string
	clientService *ClientService
}

func NewServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr:    listenAddr,
		clientService: NewClientService(),
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.Handle("/client_service", s.clientService)
	fmt.Printf("Serving %s\n", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}
