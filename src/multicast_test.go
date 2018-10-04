package main

import (
	"net"
	"reflect"
	"testing"
	"time"
)

type packetConnStub struct {
	net.PacketConn
}

var stubChannel chan []byte

func (s packetConnStub) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	stubData := <-stubChannel
	n = len(stubData)
	b = append(b[:0], stubData...)
	return n, nil, nil
}

func TestReadPackets(t *testing.T) {

	// connection stubbing
	multicasterListenAddress = "STUBBED"
	stubChannel = make(chan []byte, 10)
	upstreamConnection = packetConnStub{}

	var (
		packetsPipe     = make(chan udpBuffer, 10)
		recycledBuffers = make(chan udpBuffer, 10)
	)

	// Allocate some memory ( 5 out of 10 )
	for i := 0; i < 5; i++ {
		recycledBuffers <- udpBuffer{}
	}

	// Read UDP packets from the upstream service
	go readPackets(
		packetsPipe,
		recycledBuffers,
	)

	var testData [10][]byte

	testData[0] = []byte("zero")
	testData[1] = []byte("one")
	testData[2] = []byte("two")
	testData[3] = []byte("three")
	testData[4] = []byte("four")
	testData[5] = []byte("five")
	testData[6] = []byte("six")
	testData[7] = []byte("seven")
	testData[8] = []byte("eight")
	testData[9] = []byte("nine")

	for _, test := range testData {
		stubChannel <- test
	}

	for i := 0; i < 10; i++ {

		select {
		case test := <-packetsPipe:
			testSlice := test.arr[:test.len]
			if !reflect.DeepEqual(testData[i], testSlice) {
				t.Errorf("incorrect packets sent through packetsPipe \"%s\" != \"%s\"",
					string(testData[i]),
					string(testSlice),
				)
				t.Fail()
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("no packets sent through packetsPipe")
			t.Fail()
		}

	}
}

func TestMulticast(t *testing.T) {

	// connection stubbing
	multicasterListenAddress = "STUBBED"
	stubChannel = make(chan []byte, 10)
	upstreamConnection = packetConnStub{}

	var (
		packetsPipe     = make(chan udpBuffer, 10)
		recycledBuffers = make(chan udpBuffer, 10)
		newAddresses    = make(chan string, 255)
		deadAddresses   = make(chan string, 255)
	)

	// Allocate some memory ( 5 out of 10 )
	for i := 0; i < 5; i++ {
		recycledBuffers <- udpBuffer{}
	}

	// Read UDP packets from the upstream service
	go readPackets(
		packetsPipe,
		recycledBuffers,
	)

	// Multicast the upstream packets to the downstream connections
	go multicast(
		packetsPipe,
		recycledBuffers,
		newAddresses,
		deadAddresses,
	)

	newAddresses <- "127.0.0.1:20100"
	newAddresses <- "127.0.0.1:20101"
	newAddresses <- "127.0.0.1:20102"

	var testData [10][]byte

	testData[0] = []byte("zero")
	testData[1] = []byte("one")
	testData[2] = []byte("two")
	testData[3] = []byte("three")
	testData[4] = []byte("four")
	testData[5] = []byte("five")
	testData[6] = []byte("six")
	testData[7] = []byte("seven")
	testData[8] = []byte("eight")
	testData[9] = []byte("nine")

	var listenConnection [3]net.PacketConn
	listenConnection[0], _ = net.ListenPacket("udp", "127.0.0.1:20100")
	listenConnection[1], _ = net.ListenPacket("udp", "127.0.0.1:20101")
	listenConnection[2], _ = net.ListenPacket("udp", "127.0.0.1:20102")

	// sleep here, if data is sent before endpoints are set up, packets may be lost
	time.Sleep(100 * time.Millisecond)

	// Send test data
	for _, test := range testData {
		stubChannel <- test
	}

	buff := make([]byte, 65507, 65507)
	for _, lConn := range listenConnection {
		for i := 0; i < 10; i++ {

			n, _, err := lConn.ReadFrom(buff)

			if err != nil {
				t.Errorf("error reading from udp connection -- %s", err)
				t.Fail()
			}

			if !reflect.DeepEqual(testData[i], buff[:n]) {
				t.Errorf("incorrect packets sent through to %s",
					string(lConn.LocalAddr().String()),
				)
				t.Fail()
			}
		}

		lConn.Close()
	}
}
