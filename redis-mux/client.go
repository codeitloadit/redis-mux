package main

import (
	"fmt"
	"log"
	"net"
)

type Client struct {
	conn       net.Conn
	redisConns []net.Conn
	msgCh      chan []byte
	errCh      chan error
	pending    bool
}

func (client *Client) handleRequest() {
	// Setup redis conns for client.
	for _, host := range hosts {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			log.Fatal(err)
		}
		client.redisConns = append(client.redisConns, conn)
	}

	// Send messages from client to redis conns.
	go func() {
		defer client.conn.Close()
		for {
			msg := make([]byte, 1024)
			n, err := client.conn.Read(msg)
			if err != nil {
				fmt.Printf("%s -> ERROR %s\n", client.conn.RemoteAddr(), err)
				return
			}
			msg = msg[:n]
			fmt.Printf("%s -> %q\n", client.conn.RemoteAddr(), string(msg))
			client.pending = true
			for _, conn := range client.redisConns {
				conn.Write(msg)
			}
		}
	}()

	// Send messages from redis conns to client.
	go func() {
		for _, conn := range client.redisConns {
			go func(conn net.Conn) {
				for {
					msg := make([]byte, 1024)
					n, err := conn.Read(msg)
					if err != nil {
						client.errCh <- err
						return
					}
					client.msgCh <- msg[:n]
				}
			}(conn)
		}

		for {
			select {
			case msg := <-client.msgCh:
				fmt.Printf("%s <- %q\n", client.conn.RemoteAddr(), string(msg))
				if client.pending {
					client.conn.Write(msg)
					client.pending = false
				}
			case err := <-client.errCh:
				fmt.Printf("%s <- ERROR %s\n", client.conn.RemoteAddr(), err)
				return
			}
		}
	}()
}
