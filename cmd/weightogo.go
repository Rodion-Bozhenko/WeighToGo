package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
	"weightogo/configparser"
	"weightogo/healthcheck"
	"weightogo/loadbalancer"
	"weightogo/logger"
)

func startServers(servers []loadbalancer.Server, wg *sync.WaitGroup) {
	for _, s := range servers {
		wg.Add(1)
		go func(address string) {
			defer wg.Done()
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				respBody := fmt.Sprintf("Hello from %s", s.Address)
				w.Write([]byte(respBody))
			})

			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
			})

			http.ListenAndServe(address, mux)
		}(s.Address)
	}
}

func main() {
	config, err := configparser.ParseConfig()
	if err != nil {
		logger.Logger.Error("Config is not valid.", "err", err)
		os.Exit(1)
	}

	servers := make([]loadbalancer.Server, 0, len(config.BackendServers))
	for _, s := range config.BackendServers {
		servers = append(servers, loadbalancer.Server{
			Address:     s.Address,
			Weight:      s.Weight,
			HC_Endpoint: s.HC_Endpoint,
			HC_Interval: s.HC_Interval,
		})
	}

	var wg sync.WaitGroup
	go startServers(servers, &wg)
	wg.Wait()

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

	aliveServers := getHealthyServers(servers)

	if len(aliveServers) == 0 {
		logger.Logger.Error("No available servers. Health check failed on each server.")
		os.Exit(0)
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
			address := "http://" + s.Address + s.HC_Endpoint
			alive, err := healthcheck.IsAlive(address, time.Second*5)
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
