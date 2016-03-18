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

var conns []net.Conn

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

func handleConnection(hostConn net.Conn) {
	for _, host := range hosts {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			log.Fatal(err)
		}

		conns = append(conns, conn)
	}

	go sendFromHostToClients(hostConn)
}

func sendFromHostToClients(hostConn net.Conn) {
	ch := make(chan []byte)
	eCh := make(chan error)

	go func() {
		for {
			data := make([]byte, 1024)
			n, err := hostConn.Read(data)
			data = data[:n]
			if err != nil {
				eCh <- err
				return
			}
			ch <- data
		}
	}()

	for {
		select {
		case data := <-ch:
			fmt.Print("Sending:\t")
			fmt.Printf("%q\n", string(data))
			for _, conn := range conns {
				conn.Write(data)
				go sendFromClientsToHost(hostConn)
			}
		case err := <-eCh:
			fmt.Println(err)
			break
		}
	}
}

func sendFromClientsToHost(hostConn net.Conn) {
	for _, conn := range conns {
		ch := make(chan []byte)
		eCh := make(chan error)

		go func() {
			for {
				data := make([]byte, 1024)
				n, err := conn.Read(data)
				data = data[:n]
				if err != nil {
					eCh <- err
					return
				}
				ch <- data
			}
		}()

		for {
			select {
			case data := <-ch:
				fmt.Print("Received:\t")
				fmt.Printf("%q\n", string(data))
				hostConn.Write(data)
			case err := <-eCh:
				fmt.Println(err)
				break
			}
		}
	}

}
