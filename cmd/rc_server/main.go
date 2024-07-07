package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/CGSG-2021-AE4/rc2/api/server"
)

func main() {
	fmt.Println("CGSG forever!!!")

	// Arguments
	host := flag.String("host", "localhost", "Host name")
	httpPort := flag.String("http-port", "8081", "HTTP request port")
	tcpPort := flag.String("tcp-port", "3047", "TCP request port")
	statFile := flag.String("stat-file", "stat.json", "Statistic filename")
	enableLog := flag.Bool("log-state", false, "Enable state logging")
	statUpdateTimeout := flag.Int64("stat-update-timeout", 10000, "Timeout between statistics' updates in MS")
	flag.Parse()

	env := server.EnvVariables{
		Host:           *host,
		HttpPort:       *httpPort,
		TcpPort:        *tcpPort,
		EnableStateLog: *enableLog,
		StatisticsFile: *statFile,
		StatTimeout:    time.Duration(*statUpdateTimeout),
	}

	apiServer := server.New(env)

	// Interuption handling
	interupt := make(chan os.Signal, 1)
	signal.Notify(interupt, os.Interrupt)

	go func(apiServer *server.Server) {
		signal := <-interupt
		log.Println("Got interupt signal: ", signal.String())
		apiServer.Close()
		<-time.After(2 * time.Second)
	}(apiServer)

	apiServer.Run() // TODO change to RunAndLog as others

	log.Println("Main finished")
}
