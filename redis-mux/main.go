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
var clients []string

func init() {
	port = os.Getenv("REDIS_MUX_PORT")
	clients = strings.Split(os.Getenv("REDIS_MUX_CLIENTS"), ",")
}

func main() {
	redisHost := clients[0]
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
		redisConn := dialRedis(redisHost, hostConn)
		go func(c net.Conn) {
			_, err := io.Copy(redisConn, c)
			if err != nil {
				log.Fatal(err)
			}
			c.Close()
		}(hostConn)
	}
}

func dialRedis(redisHost string, hostConn net.Conn) net.Conn {
	redisConn, err := net.Dial("tcp", redisHost)
	if err != nil {
		log.Fatal(err)
	}
	go func(c net.Conn) {
		for {
			reader := bufio.NewReader(c)
			_, err := io.Copy(hostConn, reader)
			if err != nil {
				log.Fatal(err)
			}
		}
	}(redisConn)
	return redisConn
}
