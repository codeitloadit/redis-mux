package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
)

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

func main() {
	port := os.Args[1]
	redisHost := os.Args[2]
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
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
