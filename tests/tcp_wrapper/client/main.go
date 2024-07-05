package main

import (
	"context"
	"log"
	tcpw "rc/api/tcp_wrapper"
	"time"
)

// Just for eazier error handling
func run() error {
	dialerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connectCtx, cancel := context.WithTimeout(dialerCtx, 10*time.Second)
	defer cancel()

	server := tcpw.ServerDescriptor{Raddr: "localhost:3044", AuthMsg: []byte("HI")}
	log.Println("Connecting...")
	conn, err := server.Connect(dialerCtx, connectCtx)
	if err != nil {
		return err
	}
	log.Println("CONNNNNNNNNNNNNNNNNECTEEEEEEEEEEEEED")
	for {
		log.Println("Reading...")
		readCtx, cancel := context.WithTimeout(dialerCtx, 4*time.Second)
		_ = cancel // AAAA there is no other way...
		msgType, msg, err := conn.Read(readCtx)
		if err != nil {
			log.Println("Read finished with error:", err.Error())
		}
		log.Println(msgType, string(msg))
	}
}

func main() {
	log.Println("CGSG forever!!!")
	if err := run(); err != nil {
		log.Println("Run finished with error:", err.Error())
	}
	log.Println("END")
}
