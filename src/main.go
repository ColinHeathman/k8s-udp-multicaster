package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {

	// Initialization
	configured :=
		configureMulticaster() &&
			configureEndpoints()
	if !configured {
		os.Exit(1)
	}

	fmt.Print("\x1b[32m[STARTING UP]\x1b[0m\n")

	// How many UDP packets to keep
	channelSize := 10
	channelSizeTxt := os.Getenv("UDP_QUEUE_MAX")
	if channelSizeTxt != "" {

		size, err := strconv.Atoi(channelSizeTxt)
		if err != nil {
			fmt.Printf("\x1b[93m[WARNG]\x1b[0m Invalid ${UDP_QUEUE_MAX} \"%s\" defaulting to 10\n",
				channelSizeTxt,
			)
		} else {
			channelSize = size
		}
	}

	var (
		packetsPipe     = make(chan udpBuffer, channelSize)
		recycledBuffers = make(chan udpBuffer, channelSize)
		newAddresses    = make(chan string, 255)
		deadAddresses   = make(chan string, 255)
	)

	// Allocate some memory
	for i := 0; i < 10; i++ {
		recycledBuffers <- udpBuffer{}
	}

	// Read and track IP addresses for the k8s service
	go maintainEndpoints(
		newAddresses,
		deadAddresses,
	)

	// Read UDP packets from the upstream service
	go readPackets(
		packetsPipe,
		recycledBuffers,
	)

	// Multicast the upstream packets to the downstream connections
	multicast(
		packetsPipe,
		recycledBuffers,
		newAddresses,
		deadAddresses,
	)
}
