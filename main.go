package main

import (
	"io"
	"log/slog"
	"net"
	"os"
	"weightogo/server"
)

var Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		Logger.Error("Error setting up listener: %v", "err", err.Error())
		listener.Close()
	}

	go server.Server()

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		Logger.Warn("Error retrieving host/port", "err", err.Error())
	}

	Logger.Info("Listening on host: %v, port: %v", host, port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			Logger.Error("Error accepting connection: %v", "err", err.Error())
			continue
		}

		go handleConnection(conn, "localhost:5000")
	}
}

func handleConnection(clientConn net.Conn, targetAddr string) {
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		Logger.Error("Error connecting to target server", "err", err.Error(), "target", targetAddr)
	}

	go func() {
		defer clientConn.Close()
		defer targetConn.Close()

		Logger.Info("Client -> Target")
		io.Copy(targetConn, clientConn)
	}()

	go func() {
		defer clientConn.Close()
		defer targetConn.Close()

		Logger.Info("Target -> Client")
		io.Copy(clientConn, targetConn)
	}()
}
