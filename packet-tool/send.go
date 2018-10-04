package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"
)

func send() {

	fmt.Print("\x1b[32m[STARTING UP]\x1b[0m\n")

	protocol := os.Getenv("PROTOCOL")
	hostname := os.Getenv("DIAL_HOST")

	ips, err := net.LookupIP(hostname)
	if err != nil {
		fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
		awaitInterrupt()
	}
	if len(ips) == 0 {
		fmt.Printf("\x1b[31m[FATAL]\x1b[0m no addresses found for: %s\n", hostname)
		awaitInterrupt()
	}

	address := ips[0].String() + ":" + os.Getenv("PORT")

	fmt.Printf("\x1b[34m sending - \x1b[0m%s@%s\n",
		protocol,
		address,
	)

	conn, err := net.Dial(protocol, address)
	if err != nil {
		fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
		awaitInterrupt()
	}

	myName := os.Getenv("POD_NAME")
	myIP := os.Getenv("POD_IP")
	nodeName := os.Getenv("NODE_NAME")

	message := fmt.Sprintf("%s @%s, from %s",
		myName,
		myIP,
		nodeName,
	)

	data, err := ioutil.ReadAll(
		strings.NewReader(message),
	)
	if err != nil {
		fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
		awaitInterrupt()
	}

	for {
		n, err := conn.Write(data)

		fmt.Printf("\x1b[34m sent - \x1b[0m%s@%s - \"%s\"\n",
			protocol,
			address,
			string(data[:n]),
		)

		if err != nil {
			fmt.Printf("\x1b[33m[ERROR]\x1b[0m, %s\n", err)
		}
		time.Sleep(time.Duration(1) * time.Second)
	}

}
