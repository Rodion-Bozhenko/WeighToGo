package main

import (
	"bytes"
	"net"
	"testing"
	"time"
)

func TestIntegration(t *testing.T) {
	// Server that echoes back whatever it receives
	echoServer := func(addr string) {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			t.Fatalf("Error setting up echo server: %v", err)
			return
		}
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				t.Logf("Echo server accept error: %v", err)
				return
			}

			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						return
					}
					c.Write(buf[:n])
				}
			}(conn)
		}
	}

	for _, addr := range []string{"localhost:5000", "localhost:5001", "localhost:5002"} {
		go echoServer(addr)
	}

	go main()

	// Giving some time for servers and the main listener to start
	time.Sleep(2 * time.Second)

	// Connect to the main application and send/receive data
	testData := []byte("test_data")
	for i := 0; i < 10; i++ {
		conn, err := net.Dial("tcp", "localhost:8080")
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close()

		conn.Write(testData)

		buf := make([]byte, len(testData))
		_, err = conn.Read(buf)
		if err != nil {
			t.Fatalf("Failed to read from connection: %v", err)
		}

		if !bytes.Equal(buf, testData) {
			t.Fatalf("Expected %s, got %s", testData, buf)
		}
	}
}
