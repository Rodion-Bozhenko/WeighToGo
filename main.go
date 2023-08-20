package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"weightogo/server"
)

var Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

type lbStrat interface {
	pickServer() string
}

type roundRobin struct {
	count     int
	addresses []string
}

func (rb *roundRobin) pickServer() string {
	index := rb.count % len(rb.addresses)
	rb.count++
	return rb.addresses[index]
}

func pickServer(strat lbStrat) string {
	return strat.pickServer()
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		Logger.Error("Error setting up listener: %v", "err", err.Error())
		listener.Close()
	}

	servers := []string{"localhost:5000", "localhost:5001", "localhost:5002"}
	for _, addr := range servers {
		addr := addr
		go server.Server(addr)
	}

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		Logger.Warn("Error retrieving host/port", "err", err.Error())
	}

	Logger.Info(fmt.Sprintf("Listening on host: %v, port: %v", host, port))

	handleConnection(listener, servers)
}

func handleConnection(listener net.Listener, servers []string) {
	roundRobinStrat := &roundRobin{addresses: servers}

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			Logger.Error("Error accepting connection: %v", "err", err.Error())
			continue
		}

		targetAddr := pickServer(roundRobinStrat)

		go func() {
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
		}()
	}
}
