package commons

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestListenTCP(t *testing.T) {
	listeningPort, err := net.Listen("tcp", "127.0.0.1:0")
	defer listeningPort.Close()
	if err != nil {
		t.Fatal(err)
	}

	port := listeningPort.Addr()
	effectivePort := strings.Split(port.String(), ":")[1]
	WaitForPort("tcp", "127.0.0.1", effectivePort, 5*time.Second)
}

func TestListenUDP(t *testing.T) {
	listeningPort, err := net.Listen("tcp", "127.0.0.1:0")
	defer listeningPort.Close()
	if err != nil {
		t.Fatal(err)
	}

	port := listeningPort.Addr()
	effectivePort := strings.Split(port.String(), ":")[1]
	WaitForPort("udp", "127.0.0.1", effectivePort, 5*time.Second)
}
