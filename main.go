package main

import (
	"flag"
	"fmt"
	"rc/api"
)

func main() {
	fmt.Println("CGSG forever!!!")

	host := flag.String("h", "localhost", "Host name")
	flag.Parse()

	server := api.NewServer(*host, "8080", "3047")
	server.Run()
}
