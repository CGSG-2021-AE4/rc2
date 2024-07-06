package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/CGSG-2021-AE4/rc2/api/server/client_service"
	msg_service "github.com/CGSG-2021-AE4/rc2/api/server/message_service"
	"github.com/CGSG-2021-AE4/rc2/api/server/statistic_service"
)

// All environment variables that can be set with start arguments
type EnvVariables struct {
	Host           string
	HttpPort       string
	TcpPort        string
	EnableStateLog bool
	StatisticsFile string
	StatTimeout    time.Duration // In MS
}

// Main API server
type Server struct {
	env               EnvVariables
	httpServer        *http.Server
	clientService     *client_service.Service
	msgHandlerService *msg_service.Service
	statService       *statistic_service.Service
	DoneChan          chan struct{} // Chanel that make all other threads like stat service stop
}

func HandleF(hc interface{ HandleHTTP(c *gin.Context) }) gin.HandlerFunc {
	return func(c *gin.Context) {
		hc.HandleHTTP(c)
	}
}

func New(env EnvVariables) *Server {
	apiPtr := &Server{
		env:      env,
		DoneChan: make(chan struct{}),
	}

	// Create http server

	// Create services
	apiPtr.statService = statistic_service.New(env.StatisticsFile)
	apiPtr.clientService = client_service.New(env.Host + ":" + env.TcpPort)
	apiPtr.msgHandlerService = msg_service.New(apiPtr.clientService)
	return apiPtr
}

func (s *Server) Run() {
	// Start services
	s.clientService.Run()
	s.statService.Run()

	// Run http services
	router := gin.New()
	router.POST("/send", HandleF(s.msgHandlerService))

	s.httpServer = &http.Server{Addr: s.env.Host + ":" + s.env.HttpPort, Handler: router}

	log.Printf("Serving HTTP %s\n", s.env.Host+":"+s.env.HttpPort)
	if err := s.httpServer.ListenAndServe(); err != nil {
		log.Println(err.Error())
	}
	log.Println("END")
}

func (s *Server) Close() {
	log.Println("BBBB")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	close(s.DoneChan)
	s.httpServer.Shutdown(ctx)
	if err := s.clientService.Close(); err != nil {
		log.Println("Client service closed with error:", err)
	}
}
