package main

import (
	"fmt"
	"net"
	"sync"
	"time"
	"weightogo/healthcheck"
	"weightogo/loadbalancer"
	"weightogo/logger"
)

var servers = []loadbalancer.Server{
	{Address: "http://localhost:7230"},
	{Address: "http://localhost:9001"},
	{Address: "http://localhost:8002"},
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		logger.Logger.Error("Error setting up listener", "err", err.Error())
		listener.Close()
	}
	defer listener.Close()

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		logger.Logger.Warn("Error retrieving host/port", "err", err.Error())
	}

	logger.Logger.Info(fmt.Sprintf("Listening on host: %v, port: %v", host, port))

	aliveServers := getHealthyServers(servers)

	if len(aliveServers) == 0 {
		logger.Logger.Error("No available servers. Health check failed on each server.")
		return
	}
	lb := loadbalancer.GetLoadBalancer(loadbalancer.LeastConnections, aliveServers)
	connectionHandler := NewConnectionHandler(listener, lb, logger.Logger)

	connectionHandler.HandleConnection()
}

func getHealthyServers(servers []loadbalancer.Server) []loadbalancer.Server {
	ch := make(chan loadbalancer.Server, len(servers))
	var wg sync.WaitGroup

	for _, s := range servers {
		wg.Add(1)
		go func(s loadbalancer.Server) {
			defer wg.Done()
			alive, err := healthcheck.IsAlive(s.Address, time.Second*5)
			if err != nil {
				logger.Logger.Error(fmt.Sprintf("Unable to healthcheck server %s", s.Address), "err", err)
			}
			if alive {
				ch <- s
			}
		}(s)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var aliveServers []loadbalancer.Server
	for s := range ch {
		aliveServers = append(aliveServers, s)
	}

	return aliveServers
}
