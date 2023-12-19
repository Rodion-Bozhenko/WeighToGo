package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
	"weightogo/configparser"
	"weightogo/healthcheck"
	"weightogo/loadbalancer"
	"weightogo/logger"
)

// ServerWithStatus shows if Server is alive
type ServerWithStatus struct {
	Server loadbalancer.Server
	Alive  bool
}

func main() {
	config, err := configparser.ParseConfig("")
	if err != nil {
		logger.Logger.Error("Config is not valid.", "err", err)
		os.Exit(1)
	}

	servers := parseServers(config.BackendServers)

	listener, err := net.Listen("tcp", config.General.BindAddress)
	if err != nil {
		logger.Logger.Error("Error setting up listener", "err", err.Error())
		listener.Close()
		os.Exit(1)
	}
	defer listener.Close()

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		logger.Logger.Warn("Error retrieving host/port", "err", err.Error())
	}

	logger.Logger.Info(fmt.Sprintf("Listening on host: %v, port: %v", host, port))

	initializeServersStatus(servers)

	var c int
	for _, s := range servers {
		if !s.Alive {
			c++
		}
	}
	if len(servers) == c {
		logger.Logger.Error("No available servers. Health check failed on each server.")
		os.Exit(0)
	}

	lb := loadbalancer.GetLoadBalancer(config.Strategy, servers)
	connectionHandler := NewConnectionHandler(listener, lb, logger.Logger)

	go healthCheckServers(servers)

	connectionHandler.HandleConnection()
}

func parseServers(backendServers []configparser.BackendServer) []*loadbalancer.Server {
	servers := make([]*loadbalancer.Server, 0, len(backendServers))
	for _, s := range backendServers {
		servers = append(servers, &loadbalancer.Server{
			Address:    s.Address,
			Weight:     s.Weight,
			HCEndpoint: s.HCEndpoint,
			HCInterval: s.HCInterval,
			Alive:      false,
		})
	}

	return servers
}

func initializeServersStatus(servers []*loadbalancer.Server) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, s := range servers {
		wg.Add(1)
		go func(s *loadbalancer.Server) {
			defer wg.Done()

			address := "http://" + s.Address + s.HCEndpoint
			alive, err := healthcheck.IsAlive(address, time.Second*5)
			if err != nil {
				logger.Logger.Warn(fmt.Sprintf("Unable to healthcheck server %s", s.Address), "err", err)
			}
			if alive {
				mu.Lock()
				s.Alive = true
				mu.Unlock()
			}
		}(s)
	}

	wg.Wait()
}

func healthCheckServers(servers []*loadbalancer.Server) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, s := range servers {
		wg.Add(1)
		go func(s *loadbalancer.Server) {
			defer wg.Done()

			ticker := time.NewTicker(s.HCInterval)
			quit := make(chan struct{})
			for {
				select {
				case <-ticker.C:
					address := "http://" + s.Address + s.HCEndpoint
					alive, err := healthcheck.IsAlive(address, time.Second*5)
					if err != nil {
						logger.Logger.Warn(fmt.Sprintf("Unable to healthcheck server %s", s.Address), "err", err)
						if s.Alive {
							mu.Lock()
							s.Alive = false
							mu.Unlock()
						}
						close(quit)
					}
					if !s.Alive && alive {
						mu.Lock()
						s.Alive = true
						mu.Unlock()
					} else if s.Alive && !alive {
						mu.Lock()
						s.Alive = false
						mu.Unlock()
					}
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}(s)
	}

	wg.Wait()
}
