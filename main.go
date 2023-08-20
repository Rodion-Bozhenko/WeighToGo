package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	serv "weightogo/server"
)

var Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

type server struct {
	address           string
	weight            int
	currentWeight     int
	activeConnections int64
}

type loadBalancer interface {
	pickServer() *server
}

type Strategy string

const (
	RoundRobin         Strategy = "RoundRobin"
	WeightedRoundRobin Strategy = "WeightedRoundRobin"
	LeastConnections   Strategy = "LeastConnections"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		Logger.Error("Error setting up listener", "err", err.Error())
		listener.Close()
	}

	servers := []server{
		{address: "localhost:5000", weight: 1},
		{address: "localhost:5001", weight: 5},
		{address: "localhost:5002", weight: 1},
	}

	for _, server := range servers {
		addr := server.address
		go serv.Server(addr)
	}

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		Logger.Warn("Error retrieving host/port", "err", err.Error())
	}

	Logger.Info(fmt.Sprintf("Listening on host: %v, port: %v", host, port))

	handleConnection(listener, servers, LeastConnections)
}

func handleConnection(listener net.Listener, servers []server, strategy Strategy) {
	loadBalancer := getLoadBalancer(strategy, servers)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			Logger.Error("Error accepting connection", "err", err.Error())
			continue
		}

		targetServer := loadBalancer.pickServer()
		targetAddr := targetServer.address
		targetServer.increaseConnections()

		go func() {
			targetConn, err := net.Dial("tcp", targetAddr)
			if err != nil {
				Logger.Error("Error connecting to target server", "err", err.Error(), "target", targetAddr)
				return
			}
			defer targetServer.decreaseConnections()

			go func() {
				defer clientConn.Close()
				defer targetConn.Close()

				io.Copy(targetConn, clientConn)
			}()

			go func() {
				defer clientConn.Close()
				defer targetConn.Close()

				io.Copy(clientConn, targetConn)
			}()
		}()
	}
}

func getLoadBalancer(strategy Strategy, servers []server) loadBalancer {
	switch strategy {
	case RoundRobin:
		return &roundRobinLoadBalancer{servers: servers, count: 0}
	case WeightedRoundRobin:
		return &smoothedLoadBalancer{servers: servers}
	case LeastConnections:
		return &leastConnectionsLoadBalancer{servers: servers}
	default:
		return &roundRobinLoadBalancer{servers: servers, count: 0}
	}
}
