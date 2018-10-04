package main

import (
	"context"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAddEndpoints(t *testing.T) {

	// Use a timeout to keep the test from hanging.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	k8sInterface = kubernetes.Interface(fake.NewSimpleClientset())
	namespace = "test-ns"
	service = "test-service"
	servicePort = "udp"
	syncInterval, _ = time.ParseDuration("100ms")

	var (
		newAddresses  = make(chan string, 255)
		deadAddresses = make(chan string, 255)
	)

	go maintainEndpoints(
		newAddresses,
		deadAddresses,
	)

	// Inject an event into the fake client.
	ep := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Name: service},
		Subsets: []v1.EndpointSubset{
			v1.EndpointSubset{
				Addresses: []v1.EndpointAddress{
					v1.EndpointAddress{
						Hostname: "test-host-00",
						IP:       "192.168.0.10",
					},
					v1.EndpointAddress{
						Hostname: "test-host-01",
						IP:       "192.168.0.11",
					},
					v1.EndpointAddress{
						Hostname: "test-host-02",
						IP:       "192.168.0.12",
					},
					v1.EndpointAddress{
						Hostname: "test-host-03",
						IP:       "192.168.0.13",
					},
				},
				Ports: []v1.EndpointPort{
					v1.EndpointPort{
						Name:     "http",
						Port:     80,
						Protocol: "TCP",
					},
					v1.EndpointPort{
						Name:     "udp",
						Port:     8462,
						Protocol: "UDP",
					},
				},
			},
		},
	}
	_, err := k8sInterface.Core().Endpoints(namespace).Create(ep)
	if err != nil {
		t.Errorf("error injecting endpoints add: %v", err)
	}

	expectedAddresses := make(map[string]bool)

	expectedAddresses["192.168.0.10:8462"] = true
	expectedAddresses["192.168.0.11:8462"] = true
	expectedAddresses["192.168.0.12:8462"] = true
	expectedAddresses["192.168.0.13:8462"] = true

	// Wait for tests to finish
	<-ctx.Done()
loop:
	for i := 0; i < 4; i++ {
		select {
		case address := <-newAddresses:
			if !expectedAddresses[address] {
				t.Errorf("maintainEndpoints sent the wrong new addresses %s", address)
				t.Fail()
			}
			t.Logf("Got address from channel: %s", address)

		default:
			t.Error("maintainEndpoints did not send new addresses")
			t.Fail()
			break loop
		}
	}
}

func TestRemoveEndpoints(t *testing.T) {

	// Use a timeout to keep the test from hanging.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	k8sInterface = kubernetes.Interface(fake.NewSimpleClientset())
	namespace = "test-ns"
	service = "test-service"
	servicePort = "udp"
	syncInterval, _ = time.ParseDuration("1ms")

	var (
		newAddresses  = make(chan string, 255)
		deadAddresses = make(chan string, 255)
	)

	go maintainEndpoints(
		newAddresses,
		deadAddresses,
	)

	// Inject an event into the fake client.
	ep := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Name: service},
		Subsets: []v1.EndpointSubset{
			v1.EndpointSubset{
				Addresses: []v1.EndpointAddress{
					v1.EndpointAddress{
						Hostname: "test-host-00",
						IP:       "192.168.0.10",
					},
					v1.EndpointAddress{
						Hostname: "test-host-01",
						IP:       "192.168.0.11",
					},
					v1.EndpointAddress{
						Hostname: "test-host-02",
						IP:       "192.168.0.12",
					},
					v1.EndpointAddress{
						Hostname: "test-host-03",
						IP:       "192.168.0.13",
					},
				},
				Ports: []v1.EndpointPort{
					v1.EndpointPort{
						Name:     "http",
						Port:     80,
						Protocol: "TCP",
					},
					v1.EndpointPort{
						Name:     "udp",
						Port:     8462,
						Protocol: "UDP",
					},
				},
			},
		},
	}
	_, err := k8sInterface.Core().Endpoints(namespace).Create(ep)
	if err != nil {
		t.Errorf("error injecting endpoints add: %v", err)
	}

	expectedAddresses := make(map[string]bool)

	expectedAddresses["192.168.0.10:8462"] = true
	expectedAddresses["192.168.0.11:8462"] = true
	expectedAddresses["192.168.0.12:8462"] = true
	expectedAddresses["192.168.0.13:8462"] = true
loop:
	for i := 0; i < 4; i++ {
		select {
		case address := <-newAddresses:
			if !expectedAddresses[address] {
				t.Errorf("maintainEndpoints sent the wrong new addresses %s", address)
				t.Fail()
			}

		case <-time.After(1 * time.Second):
			t.Error("maintainEndpoints did not send new addresses")
			t.Fail()
			break loop
		}
	}

	// Modify the endpoints
	ep2 := &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{Name: service},
		Subsets: []v1.EndpointSubset{
			v1.EndpointSubset{
				Addresses: []v1.EndpointAddress{
					v1.EndpointAddress{
						Hostname: "test-host-00",
						IP:       "192.168.0.10",
					},
					v1.EndpointAddress{
						Hostname: "test-host-01",
						IP:       "192.168.0.11",
					},
					v1.EndpointAddress{
						Hostname: "test-host-03",
						IP:       "192.168.0.13",
					},
				},
				Ports: []v1.EndpointPort{
					v1.EndpointPort{
						Name:     "http",
						Port:     80,
						Protocol: "TCP",
					},
					v1.EndpointPort{
						Name:     "udp",
						Port:     8462,
						Protocol: "UDP",
					},
				},
			},
		},
	}
	_, err = k8sInterface.Core().Endpoints(namespace).Update(ep2)
	if err != nil {
		t.Errorf("error injecting endpoints2 add: %v", err)
	}

	expectedRemovedAddresses := make(map[string]bool)
	expectedRemovedAddresses["192.168.0.12:8462"] = true

	// Wait for tests to finish
	<-ctx.Done()
loop2:
	for i := 0; i < 1; i++ {
		select {
		case address := <-deadAddresses:
			if !expectedRemovedAddresses[address] {
				t.Errorf("maintainEndpoints sent the wrong dead addresses %s", address)
				t.Fail()
			}
			t.Logf("Got dead address from channel: %s", address)

		default:
			t.Error("maintainEndpoints did not send dead addresses")
			t.Fail()
			break loop2
		}
	}
}
