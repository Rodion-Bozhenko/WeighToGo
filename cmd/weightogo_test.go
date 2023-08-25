package main

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
	"weightogo/loadbalancer"
)

var servers = []loadbalancer.Server{
	{Address: "http://localhost:7230"},
	{Address: "http://localhost:9001"},
	{Address: "http://localhost:8002"},
}

func TestIntegration(t *testing.T) {
	startEchoServers(t, servers)
	go main()

	// Time for servers to start
	time.Sleep(2 * time.Second)

	testData := []byte("test_data")
	for i := 0; i < 10; i++ {
		testLoadBalancer(t, "localhost:8080", testData)
	}
}

func TestHealthcheckIntegration(t *testing.T) {
	// Not running third server
	startEchoServers(t, servers[:2])
	time.Sleep(1 * time.Second)

	go main()

	// Time for servers to start
	time.Sleep(2 * time.Second)

	healthyServers := getHealthyServers(servers)
	if len(healthyServers) != 2 {
		t.Fatalf("Expected 2 healthy servers, got %d", len(healthyServers))
	}
	for _, server := range healthyServers {
		if server.Address == "http://localhost:8002" {
			t.Fatal("localhost:8002 should not be marked as healthy")
		}
	}
}

func startEchoServers(t *testing.T, servers []loadbalancer.Server) {
	var wg sync.WaitGroup
	errChan := make(chan error)
	echoServer := func(addr string) {
		parsedUrl, _ := url.Parse(addr)
		host := parsedUrl.Hostname()
		port := parsedUrl.Port()
		address := fmt.Sprintf("%s:%s", host, port)

		listener, err := net.Listen("tcp", address)
		if err != nil {
			errChan <- err
			return
		}
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Logf("Echo server accept error: %v", err)
				return
			}

			wg.Add(1)
			go func(conn net.Conn) {
				defer wg.Done()
				defer conn.Close()
				buf := make([]byte, 1024)
				for {
					n, err := conn.Read(buf)
					if err != nil {
						return
					}
					responseBody := string(buf)
					responseHeaders := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
						"Content-Length: %d\r\n"+
						"Content-Type: text/plain\r\n"+
						"\r\n", len(responseBody))

					conn.Write([]byte(responseHeaders + responseBody[:n]))
				}
			}(conn)
		}
	}

	for _, s := range servers {
		go echoServer(s.Address)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			t.Fatalf("Error setting up echo server: %v", err)
		}
	}
}

func testLoadBalancer(t *testing.T, address string, testData []byte) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	conn.Write(testData)

	buf := make([]byte, 1000)
	_, err = conn.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read from connection: %v", err)
	}

	resp := strings.Split(string(buf), "\n")
	respBody := resp[len(resp)-1]

	if strings.Compare(respBody, string(testData)) == 0 {
		t.Fatalf("Expected %s, got %s", testData, respBody)
	}
}
