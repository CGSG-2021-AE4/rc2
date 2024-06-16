package main

import (
	"flag"
	"fmt"
	"rc/api"
)

func main() {
	fmt.Println("CGSG forever!!!")

	host := flag.String("host", "localhost", "Host name")
	httpPort := flag.String("http-port", "8080", "HTTP request port")
	tcpPort := flag.String("tcp-port", "3047", "TCP request port")
	flag.Parse()

	server := api.NewServer(*host, *httpPort, *tcpPort)
	server.Run()
}
