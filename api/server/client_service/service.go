package client_service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/CGSG-2021-AE4/rc2/api"
	tcpw "github.com/CGSG-2021-AE4/rc2/pkg/tcp_wrapper"
)

/* The best way to shut down service I think is through a context
 * It will go through all vertical stuff
 */

type Service struct {
	listenAddr string

	listenerContext context.Context // Used to shut down listenning
	stopListener    context.CancelFunc

	clients *clientRegister
}

// Constructor
func New(listenAddr string) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		listenAddr:      listenAddr,
		listenerContext: ctx,
		stopListener:    cancel,
		clients:         newClientRegister(),
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
		connAc, err := server.Listen(cs.listenerContext)
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
			connAc.Decline(cs.listenerContext, []byte("Client with login "+authMsg.Login+" is already connected"))
			continue ListenLoop
		}

		// Accepting
		conn, err := connAc.Accept(cs.listenerContext)
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
	go api.RunAndLog(cs.RunSync, "client service")
}

func (cs *Service) Close() {
	cs.stopListener() // Cancel context
}
