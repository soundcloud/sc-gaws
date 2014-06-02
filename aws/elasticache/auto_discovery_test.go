package elasticache

import (
	"bufio"
	"fmt"
	"net"
	"reflect"
	"testing"
)

const (
	serverListString = "test-15em9casuzor3.uq0fwm.0001.use1.cache.amazonaws.com|192.0.2.1|11211 test-15em9casuzor3.uq0fwm.0002.use1.cache.amazonaws.com|192.0.2.2|11211 test-15em9casuzor3.uq0fwm.0003.use1.cache.amazonaws.com|192.0.2.3|11211"
)

var (
	serverListStruct = []ElastiCacheServer{
		ElastiCacheServer{
			"test-15em9casuzor3.uq0fwm.0001.use1.cache.amazonaws.com",
			"192.0.2.1",
			11211,
		},
		ElastiCacheServer{
			"test-15em9casuzor3.uq0fwm.0002.use1.cache.amazonaws.com",
			"192.0.2.2",
			11211,
		},
		ElastiCacheServer{
			"test-15em9casuzor3.uq0fwm.0003.use1.cache.amazonaws.com",
			"192.0.2.3",
			11211,
		},
	}
)

func server(t *testing.T, addrChan chan<- string) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err.Error())
	}
	addrChan <- listener.Addr().String()

	conn, err := listener.Accept()
	t.Logf("Test listening on: %s", listener.Addr().String())
	if err != nil {
		t.Fatal(err.Error())
	}
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	line, err := rw.ReadString('\n')
	if line != "config get cluster\r\n" {
		t.Fatalf("Expected 'config get cluster' command, but got: %v", []byte(line))
	}

	// Write out fake cluster configuration
	rw.WriteString("CONFIG cluster 0 218\r\n")
	rw.WriteString("1\n")
	rw.WriteString(fmt.Sprintf("%s\n\r\n", serverListString))
	rw.WriteString("END\r\n")
	rw.Flush()

	conn.Close()
}

func TestConfigOutputSchema(t *testing.T) {
	// listener.Accept() blocks until a connection is received, so we need to
	// set the server up in a goroutine and wait for the listen port to be
	// available.
	addrChan := make(chan string)
	go server(t, addrChan)

	// Create an AutoDiscoverer
	autoDisc, err := NewAutoDiscoverer(<-addrChan)
	if err != nil {
		t.Fatal(err.Error())
	}

	version, servers, err := autoDisc.GetClusterConfig()
	if err != nil {
		t.Fatal(err.Error())
	}

	if version != 1 {
		t.Fatalf("Expected version to be 1, but it was %d", version)
	}

	if !reflect.DeepEqual(servers, serverListStruct) {
		t.Fatalf("new servers don't match, expected: %v, got: %v", serverListStruct, servers)
	}
}
