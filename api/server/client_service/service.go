package client_service

import (
	"log"
	"net"
	"sync"

	cw "github.com/CGSG-2021-AE4/go_utils/conn_wrapper"
)

type Service struct {
	listenAddr string
	listener   net.Listener
	connMutex  sync.Mutex
	Conns      map[string]*Conn // SHIT not thread safe
}

// Constructor
func New(listenAddr string) *Service {
	return &Service{
		listenAddr: listenAddr,
		Conns:      make(map[string]*Conn),
	}
}

func (cs *Service) Run() error {
	log.Println("TCP server: ", cs.listenAddr)
	listener, err := net.Listen("tcp", cs.listenAddr)
	if err != nil {
		return err
	}
	cs.listener = listener

	// Start accept cycle
	go func() (err error) {
		defer func() {
			if err != nil {
				log.Println("End accept cycle with error:", err.Error())
			} else {
				log.Println("End accept cycle")
			}
		}()

		for {
			c, err := cs.listener.Accept()
			if err != nil {
				return err
			}
			// !!! Ensure that it's right tcp connection
			// Validation msg: "TCP forever!!!"
			validationMsg := "TCP forever!!!"
			buf := make([]byte, len(validationMsg))
			if _, err := c.Read(buf); err != nil || string(buf) != validationMsg {
				log.Println("Invalid validation msg:", string(buf))
				c.Close()
				continue
			}
			log.Println("New conn: ", c.RemoteAddr().Network(), c.RemoteAddr())
			go func(c net.Conn) {
				client := NewClient(cs, cw.New(c))
				client.Run()
				log.Println("Connnection ", c.RemoteAddr().Network(), c.RemoteAddr(), " closed.")
			}(c)
		}
	}()
	return nil
}

func (cs *Service) GetCurClients() []*Conn {
	cs.connMutex.Lock()
	defer cs.connMutex.Unlock()

	conns := make([]*Conn, len(cs.Conns))
	i := 0
	for _, c := range cs.Conns {
		conns[i] = c
		i++
	}
	return conns
}

func (cs *Service) Close() error {
	if cs.listener != nil {
		log.Println("Close TCP listener")
		return cs.listener.Close()
	}
	log.Println("Close TCP clistener is NIL")
	return nil
}
