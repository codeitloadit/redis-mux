package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var port string
var hosts []string

func init() {
	port = os.Getenv("REDIS_MUX_PORT")
	hosts = strings.Split(os.Getenv("REDIS_MUX_CLIENTS"), ",")
}

func main() {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	fmt.Println("Listening on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		client := Client{conn, make([]net.Conn, 0), make(chan []byte), make(chan error), false}
		client.handleRequest()
	}
}
