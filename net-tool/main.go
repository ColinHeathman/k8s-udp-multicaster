package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {

	appType := os.Getenv("TYPE")
	protocol := os.Getenv("PROTOCOL")

	if strings.ToUpper(appType) == "SEND" {
		send()
	} else if strings.ToUpper(protocol) == "TCP" &&
		strings.ToUpper(appType) == "LISTEN" {
		tcpListen()
	} else if strings.ToUpper(protocol) == "UDP" &&
		strings.ToUpper(appType) == "LISTEN" {
		udpListen()
	} else {
		fmt.Printf("\x1b[31m[FATAL]\x1b[0m, $TYPE should be \"SEND\" or \"LISTEN\" \n")
		awaitInterrupt()
	}

}

func awaitInterrupt() {
	fmt.Printf("\x1b[32m[WAITING FOR INTERRUPT]\x1b[0m\n")

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Exit(1)
	}()

	for true {
		time.Sleep(math.MaxInt64)
	}

	os.Exit(1)
}
