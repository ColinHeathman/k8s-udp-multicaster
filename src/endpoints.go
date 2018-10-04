package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	k8sInterface kubernetes.Interface
	namespace    string
	service      string
	servicePort  string
	syncInterval time.Duration
)

func configureEndpoints() bool {

	success := true

	if k8sInterface == nil {

		// creates the in-cluster config
		config, err := rest.InClusterConfig()
		if err != nil {
			fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
			success = false
		}

		// creates the clientset
		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
			success = false
		}
		k8sInterface = kubernetes.Interface(clientSet)

	}

	if namespace == "" {

		// Get the current namespace
		namespaceFile, err := os.Open("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
			success = false
		}
		namespaceData, err := ioutil.ReadAll(namespaceFile)
		if err != nil {
			fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
			success = false
		}
		namespace = string(namespaceData)

	}

	if service == "" {

		// Get the service to multicast to
		service = os.Getenv("SERVICE_NAME")
		if service == "" {
			fmt.Print("\x1b[31m[FATAL]\x1b[0m ${SERVICE_NAME} is empty -- a k8s service name must be provided\n")
			success = false
		}

	}

	if servicePort == "" {

		// Get the service to multicast to
		servicePort = os.Getenv("SERVICE_PORT")
		if service == "" {
			fmt.Print("\x1b[31m[FATAL]\x1b[0m ${SERVICE_PORT} is empty -- a k8s service port must be provided (name or number)\n")
			success = false
		}

	}

	if syncInterval == 0 {

		// Get the k8s sync interval
		interval := os.Getenv("SYNC_INTERVAL")
		if interval == "" {
			fmt.Print("\x1b[31m[FATAL]\x1b[0m ${SYNC_INTERVAL} is empty -- a sync timeout must be provided\n")
			success = false
		}
		var err error
		syncInterval, err = time.ParseDuration(interval)
		if err != nil {
			fmt.Printf("\x1b[31m[FATAL]\x1b[0m %s\n", err)
			success = false
		}
	}

	return success
}

func maintainEndpoints(
	newAddresses chan string,
	deadAddresses chan string,
) {

	activeAddresses := make(map[string]bool)
	endpointsAddresses := make(map[string]bool)

	for {

		// Endpoints map -- this is authoritative source from kubernetes
		endpoints, err := k8sInterface.CoreV1().Endpoints(namespace).Get(service, metav1.GetOptions{})
		if err != nil {
			fmt.Printf("\x1b[33m[ERROR]\x1b[0m %s\n", err)
			time.Sleep(syncInterval)
			continue
		}

		// Clear the map
		for key := range endpointsAddresses {
			delete(endpointsAddresses, key)
		}

		// Populate the map
		for index, subset := range endpoints.Subsets {

			// Find matching port
			servicePortNumber := ""
			for _, ports := range subset.Ports {
				portNumber := strconv.Itoa(int(ports.Port))
				if ports.Name == servicePort ||
					portNumber == servicePort {
					servicePortNumber = portNumber
				}
			}
			if servicePortNumber == "" {
				fmt.Printf("\x1b[93m[WARNG]\x1b[0m No port found matching \"%s\" in endpoints %s subset %d\n",
					servicePort,
					endpoints.Name,
					index,
				)
				continue
			}

			// Compile addresses from IP & Port
			for _, addresses := range subset.Addresses {
				address := addresses.IP + ":" + servicePortNumber
				endpointsAddresses[address] = true
			}
		}

		// For each address from k8s -- look for new addresses
		for endpointAddress := range endpointsAddresses {
			if activeAddresses[endpointAddress] {
				// -- already tracking this address -- skip
				continue
			} else {
				// -- haven't seen this address before
				activeAddresses[endpointAddress] = true
				newAddresses <- endpointAddress
			}
		}

		// For each already tracked address -- look for missing addresses
		for activeAddress := range activeAddresses {
			if endpointsAddresses[activeAddress] {
				// -- this address is still active -- skip
				continue
			} else {
				// -- this address is dead
				delete(activeAddresses, activeAddress)
				deadAddresses <- activeAddress
			}
		}

		// Wait for the sync interval in time
		time.Sleep(syncInterval)
	}
}
