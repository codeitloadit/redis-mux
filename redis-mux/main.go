package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var port string
var hosts []string

type ClientData struct {
	readers []io.Reader
	writers []io.Writer
}

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

	fmt.Println("Listing on port", port)

	for {
		hostConn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		defer hostConn.Close()

		go handleConnection(hostConn)
	}
}

func handleConnection(conn net.Conn) {
	var clientData ClientData

	for _, host := range hosts {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			log.Fatal(err)
		}
		clientData.readers = append(clientData.readers, bufio.NewReader(conn))
		clientData.writers = append(clientData.writers, conn)
	}

	go sendFromHostToClients(conn, clientData)
	go sendFromClientsToHost(conn, clientData)
}

func sendFromHostToClients(hostConn net.Conn, clientData ClientData) {
	mr := io.MultiReader(clientData.readers...)
	_, err := io.Copy(hostConn, mr)
	if err != nil {
		log.Fatal(err)
	}
}

func sendFromClientsToHost(hostConn net.Conn, clientData ClientData) {
	mw := io.MultiWriter(clientData.writers...)
	_, err := io.Copy(mw, hostConn)
	if err != nil {
		log.Fatal(err)
	}
}
