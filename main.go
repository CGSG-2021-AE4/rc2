package main

import (
	"fmt"
	"rc/api"
)

func main() {
	fmt.Println("CGSG forever!!!")
	server := api.NewServer(":3047")
	server.Run()
}
