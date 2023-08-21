package main

import (
	"fmt"
	"net"
	"weightogo/loadbalancer"
	"weightogo/logger"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		logger.Logger.Error("Error setting up listener", "err", err.Error())
		listener.Close()
	}

	servers := []loadbalancer.Server{
		{Address: "localhost:5000", Weight: 1},
		{Address: "localhost:5001", Weight: 5},
		{Address: "localhost:5002", Weight: 1},
	}

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		logger.Logger.Warn("Error retrieving host/port", "err", err.Error())
	}

	logger.Logger.Info(fmt.Sprintf("Listening on host: %v, port: %v", host, port))

	lb := loadbalancer.GetLoadBalancer(loadbalancer.LeastConnections, servers)
	connectionHandler := NewConnectionHandler(listener, lb, logger.Logger)

	connectionHandler.HandleConnection()
}
