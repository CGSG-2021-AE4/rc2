package main

import (
	"context"
	"log"
	"time"

	"github.com/CGSG-2021-AE4/rc2/api"
	tcpw "github.com/CGSG-2021-AE4/rc2/pkg/tcp_wrapper"
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
	api.RunAndLog(run, "MAIN")
}
