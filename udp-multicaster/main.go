package main

import (
	"fmt"
	"net"
	"os"
)

func main() {

	fmt.Print("\x1b[32m[STARTING UP]\x1b[0m\n")

	listenAddress := os.Getenv("LISTEN_HOST") + ":" + os.Getenv("LISTEN_PORT")

	upstreamConn, err := net.ListenPacket("udp", listenAddress)
	if err != nil {
		fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("\x1b[34m listening - \x1b[0mudp@%s\n",
		listenAddress,
	)

	hostname := os.Getenv("DIAL_HOST")

	ips, err := net.LookupIP(hostname)
	if err != nil {
		fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
		os.Exit(1)
	}
	if len(ips) == 0 {
		fmt.Printf("\x1b[31m[FATAL]\x1b[0m no addresses found for: %s\n", hostname)
		os.Exit(1)
	}

	downstreamConnections := make([]net.Conn, len(ips))

	for index, ip := range ips {

		dialAddress := ip.String() + ":" + os.Getenv("DIAL_PORT")

		fmt.Printf("\x1b[34m dialing - \x1b[0mudp@%s\n",
			dialAddress,
		)

		downstreamConnections[index], err = net.Dial("udp", dialAddress)
		if err != nil {
			fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
			os.Exit(1)
		}

	}

	for {
		var buffer [1024 * 1024]byte

		n, addr, err := upstreamConn.ReadFrom(buffer[:])
		if err != nil {
			fmt.Printf("\x1b[33m[ERROR]\x1b[0m %s\n", err)
			break
		}

		fmt.Printf("\x1b[34m[INFO ]\x1b[0m multicasting packet from %s\n", addr.String())

		for _, dstrConn := range downstreamConnections {
			_, err = dstrConn.Write(buffer[:n])
			if err != nil {
				fmt.Printf("\x1b[33m[ERROR]\x1b[0m %s\n", err)
				break
			}
		}

	}

	upstreamConn.Close()

	for _, dstrConn := range downstreamConnections {
		dstrConn.Close()
	}

	fmt.Print("\x1b[32m[SHUTTING DOWN]\x1b[0m\n")
}
