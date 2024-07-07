package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/CGSG-2021-AE4/rc2/api"
	"github.com/CGSG-2021-AE4/rc2/api/client"
)

// Just for better error handling
func runSync() error {
	configFilename := flag.String("c", "./config.json", "File name of config(relative or global)")
	flag.Parse()

	// Prepare config
	config, err := client.LoadConfig(*configFilename)
	if err != nil {
		return err
	}

	// Start request service
	reqService := client.NewReqService(config)
	reqService.Run()

	// Handling interupt
	interupt := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(interupt, os.Interrupt)

	go func(rs *client.RequestService) {
		signal := <-interupt
		log.Println("Got interupt signal: ", signal.String())

		rs.Close()
		close(done)
	}(reqService)

	<-done
	return nil
}

func main() {
	fmt.Println("CGSG forever!!!")
	api.RunAndLog(runSync, "MAIN")
}
