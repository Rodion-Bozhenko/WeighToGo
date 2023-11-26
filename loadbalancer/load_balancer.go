package loadbalancer

import (
	"time"
)

type Server struct {
	Address           string
	Weight            int
	CurrentWeight     int
	ActiveConnections int64
	HCEndpoint        string
	HCInterval        time.Duration
	Alive             bool
}

// LoadBalancer provides PickServer method
type LoadBalancer interface {
	PickServer() *Server
}

// Strategy is a type alias for load balancer strategies
type Strategy string

// Constant for load balancer strategy
const (
	RoundRobin         Strategy = "RoundRobin"
	WeightedRoundRobin Strategy = "WeightedRoundRobin"
	LeastConnections   Strategy = "LeastConnections"
)

// GetLoadBalancer return pointer to load balancer corresponding to strategy
func GetLoadBalancer(strategy Strategy, servers []*Server) LoadBalancer {
	switch strategy {
	case RoundRobin:
		return &RoundRobinLoadBalancer{Servers: servers, Count: 0}
	case WeightedRoundRobin:
		return &SmoothedLoadBalancer{Servers: servers}
	case LeastConnections:
		return &LeastConnectionsLoadBalancer{Servers: servers}
	default:
		return &RoundRobinLoadBalancer{Servers: servers, Count: 0}
	}
}
