package main

import (
	"fmt"
	"net"
	"os"
)

func tcpListen() {

	fmt.Print("\x1b[32m[STARTING UP]\x1b[0m\n")

	protocol := os.Getenv("PROTOCOL")
	address := os.Getenv("LISTEN_HOST") + ":" + os.Getenv("PORT")

	fmt.Printf("\x1b[34m listening - \x1b[0m%s@%s\n",
		protocol,
		address,
	)

	ln, err := net.Listen(protocol, address)
	if err != nil {
		fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
		awaitInterrupt()
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("\x1b[33m[ERROR]\x1b[0m %s\n", err)
		}
		go handleConnection(conn)

	}

}

func handleConnection(conn net.Conn) {

	for {
		var buffer [1024]byte
		n, err := conn.Read(buffer[:])
		if err != nil {
			fmt.Printf("\x1b[33m[ERROR]\x1b[0m %s\n", err)
			break
		}

		message := string(buffer[:n])

		fmt.Printf("%s\n", message)
	}

	conn.Close()

}
