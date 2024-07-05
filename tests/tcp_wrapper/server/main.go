package main

import (
	"context"
	"log"
	tcpw "rc/api/tcp_wrapper"
	"time"
)

func handleClient(c *tcpw.Conn) error {
	for {
		pause := time.After(2 * time.Second)

		log.Println("Reading...")
		readCtx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		msgType, msg, err := c.Read(readCtx)
		if err != nil {
			log.Println("Read finished with error:", err.Error())
		}
		log.Println(msgType, string(msg))
		<-pause
	}
}

func run() error {
	server, err := tcpw.NewTcpServer("localhost:3044")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		log.Println("Listenning")
		connAc, err := server.Listen(ctx)
		if err != nil {
			return err
		}
		log.Println("CONNNNNNNNNNNNNNNNNECTEEEEEEEEEEEEED")
		log.Println(connAc.AuthMsg)
		log.Println("Accepting")
		conn, err := connAc.Accept(ctx)
		if err != nil {
			return err
		}

		go func(c *tcpw.Conn) {
			if err := handleClient(c); err != nil {
				log.Println("Handle client finished with error:", err.Error())
			}
		}(conn)
	}
}

func main() {
	log.Println("CGSG forever!!!")
	if err := run(); err != nil {
		log.Println("Run finished with error:", err.Error())
	}
	log.Println("END")
}
