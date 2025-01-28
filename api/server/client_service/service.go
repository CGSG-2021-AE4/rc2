package client_service

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/CGSG-2021-AE4/rc2/api"
	tcpw "github.com/CGSG-2021-AE4/rc2/pkg/tcp_wrapper"
)

/* The best way to shut down service I think is through a context
 * It will go through all vertical stuff
 */

type Service struct {
	listenAddr string

	runCtx   context.Context // Used to shut down listenning
	stopRun  context.CancelFunc
	ctxMutex sync.Mutex

	clients *clientRegister
}

// Constructor
func New(listenAddr string) *Service {
	return &Service{
		listenAddr: listenAddr,
		clients:    newClientRegister(),
	}
}

func (cs *Service) RunSync() error {
	log.Println("TCP server: ", cs.listenAddr)
	server, err := tcpw.NewTcpServer(cs.listenAddr)
	if err != nil {
		return err
	}

ListenLoop:
	for {
		log.Println("Listenning...") // Debug
		connAc, err := server.Listen(cs.runCtx)
		if err != nil {
			return err
		}

		log.Println("Got connection")
		// Authentification
		// Later I will add some tokens
		var authMsg api.AuthMsg
		if err := json.Unmarshal(connAc.AuthMsg, &authMsg); err != nil {
			return err
		}
		if cs.clients.Exist(authMsg.Login) {
			connAc.Decline(cs.runCtx, []byte("Client with login "+authMsg.Login+" is already connected"))
			continue ListenLoop
		}

		// Accepting
		conn, err := connAc.Accept(cs.runCtx)
		if err != nil {
			return err
		}
		client := NewClient(cs, authMsg.Login, conn)
		if err := cs.clients.Add(client); err != nil { // If something went wrong
			client.Close()
			continue ListenLoop
		}
		go func(cs *Service, c *Conn) { // Because I should remove it from client register here
			api.RunAndLog(c.RunSync, "client "+c.Login)

			if err := cs.clients.Remove(c); err != nil {
				log.Println("Unregister client with error:", err.Error())
			}
		}(cs, client)
	}
}

func (cs *Service) Run() {
	// Assume that service is not running
	cs.ctxMutex.Lock()
	defer cs.ctxMutex.Unlock()

	cs.runCtx, cs.stopRun = context.WithCancel(context.Background())
	go api.RunAndLog(cs.RunSync, "client service")
}

func (cs *Service) Stop() {
	cs.ctxMutex.Lock()
	defer cs.ctxMutex.Unlock()

	if cs.runCtx != nil {
		cs.stopRun()
		cs.runCtx = nil
		cs.stopRun = nil
	}
}

func (cs *Service) Close() {
	cs.Stop()
}
