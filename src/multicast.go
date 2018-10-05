package main

import (
	"fmt"
	"net"
	"os"
)

var (
	multicasterListenAddress string
	upstreamConnection       net.PacketConn
)

func configureMulticaster() bool {

	success := true

	// Configure the listening address
	if multicasterListenAddress == "" {
		multicasterListenAddress = os.Getenv("LISTEN_HOST") + ":" + os.Getenv("LISTEN_PORT")
	}
	if multicasterListenAddress == ":" {
		fmt.Print("\x1b[31m[FATAL]\x1b[0m ${LISTEN_PORT} is empty -- a udp port be provided for listening\n")
		success = false
	}

	// Configure the upstream connection
	if upstreamConnection == nil {

		var err error
		upstreamConnection, err = net.ListenPacket("udp", multicasterListenAddress)
		if err != nil {
			fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
			success = false
		}
	}

	return success
}

type udpBuffer struct {
	arr [65507]byte
	len int
}

func readPackets(
	packetsPipe chan udpBuffer,
	recycledBuffers chan udpBuffer,
) {

	fmt.Printf("\x1b[34m listening @ \x1b[0m%s\n",
		multicasterListenAddress,
	)

	defer upstreamConnection.Close()

	for {

		var buffer udpBuffer

		select {
		case buffer = <-recycledBuffers:
		default:
			buffer = udpBuffer{}
		}

		n, _, err := upstreamConnection.ReadFrom(buffer.arr[:])
		if err != nil {
			fmt.Printf("\x1b[33m[ERROR]\x1b[0m %s\n", err)
		}
		if n > 65507 {
			fmt.Print("\x1b[33m[FATAL]\x1b[0m got packet larger than the maximum size for IPv4 UDP \n")
			os.Exit(1)
		}
		buffer.len = n

		select {
		case packetsPipe <- buffer:
			// Buffer not full
		default:
			// Buffer full
			// Discard oldest
			<-packetsPipe
			// Enqueue latest
			packetsPipe <- buffer
		}
	}

}

func multicast(
	packetsPipe chan udpBuffer,
	recycledBuffers chan udpBuffer,
	newAddresses chan string,
	deadAddresses chan string,
) {

	var (
		downstreamChannels = make(map[string]chan []byte)
		syncChannels       = make(map[string]chan bool)
	)

	for {

		select {
		// Any new addresses?
		case addr := <-newAddresses:
			downstreamChannels[addr] = make(chan []byte)
			syncChannels[addr] = make(chan bool)
			go writePackets(
				addr,
				downstreamChannels[addr],
				syncChannels[addr],
			)

		// Any dead addresses?
		case addr := <-deadAddresses:
			close(downstreamChannels[addr])
			delete(downstreamChannels, addr)
			delete(syncChannels, addr)

			fmt.Printf("\x1b[34m removing downstream @ \x1b[0m%s\n",
				addr,
			)

		// Get packet data from packets pipe
		case packetData := <-packetsPipe:
			// Send packet data to downstream endpoints
			for _, downstreamChannel := range downstreamChannels {
				downstreamChannel <- packetData.arr[:packetData.len]
			}
			// Await all send routines to complete
			for addr, syncChannel := range syncChannels {
				_, ok := <-syncChannel
				if !ok {
					close(downstreamChannels[addr])
					delete(downstreamChannels, addr)
					delete(syncChannels, addr)
				}
			}

			select {
			// Recycle packetData
			case recycledBuffers <- packetData:
			// Garbage collect packetData
			default:
			}
		}

	}

}

func writePackets(
	address string,
	downstreamChannel chan []byte,
	syncChannel chan bool,
) {

	fmt.Printf("\x1b[34m adding downstream @ \x1b[0m%s\n",
		address,
	)

	connection, err := net.Dial("udp", address)
	if err != nil {
		fmt.Printf("\x1b[31m[ERROR]\x1b[0m %s\n", err)
		close(syncChannel)
		return
	}

	defer connection.Close()

	for {
		select {
		case packetData, ok := <-downstreamChannel:
			{
				if !ok {
					fmt.Printf("\x1b[34m closing downstream @ \x1b[0m%s\n",
						address,
					)
					close(syncChannel)
					return
				}
				_, err = connection.Write(packetData)
				if err != nil {
					fmt.Printf("\x1b[33m[ERROR]\x1b[0m %s\n", err)
				}
				syncChannel <- true
			}
		}
	}
}
