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
var clientData ClientData

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

		for _, host := range hosts {
			conn, err := net.Dial("tcp", host)
			if err != nil {
				log.Fatal(err)
			}
			clientData.readers = append(clientData.readers, bufio.NewReader(conn))
			clientData.writers = append(clientData.writers, conn)
		}

		go func() {
			for {
				mr := io.MultiReader(clientData.readers...)
				_, err := io.Copy(hostConn, mr)
				if err != nil {
					log.Fatal(err)
				}
			}
		}()

		go func() {
			mw := io.MultiWriter(clientData.writers...)
			_, err := io.Copy(mw, hostConn)
			if err != nil {
				log.Fatal(err)
			}
			hostConn.Close()
		}()
	}
}
