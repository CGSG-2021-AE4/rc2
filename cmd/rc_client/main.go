package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/CGSG-2021-AE4/rc2/api/client"
)

func main() {
	fmt.Println("CGSG forever!!!")

	var configPath = flag.String("c", "./config.json", "File name of config(relative or global)")
	flag.Parse()
	var globalCS atomic.Pointer[client.Conn]

	// Handling interupt
	interupt := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(interupt, os.Interrupt)

	go func() {
		signal := <-interupt
		log.Println("Got interupt signal: ", signal.String())
		cs := globalCS.Load()
		if cs != nil {
			log.Println("AAAA")
			if err := cs.Close(); err != nil {
				log.Println(err.Error())
			}
		}
		close(done)
	}()

	reconnecting := false

	for {
		if reconnecting {
			log.Println("Reconnecting...")
			<-time.After(2 * time.Second)
		}
		select {
		case <-done:
			os.Exit(0)
		default:
			cs, err := client.NewConnection(*configPath)
			if err != nil {
				log.Println("Server connection error:", err.Error())
			}
			globalCS.Store(cs)

			// Main loop
			if err := cs.Run(); err != nil {
				log.Println("RUN finished with ERROR:", err)
			}
			globalCS.Store(nil)

			reconnecting = true // TODO remove reconnect print after interupt
		}
	}
}
