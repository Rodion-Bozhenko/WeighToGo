package loadbalancer

import (
	"time"
)

type Server struct {
	Address           string
	Weight            int
	CurrentWeight     int
	ActiveConnections int64
	HC_Endpoint       string
	HC_Interval       time.Duration
}

type LoadBalancer interface {
	PickServer() *Server
}

type Strategy string

const (
	RoundRobin         Strategy = "RoundRobin"
	WeightedRoundRobin Strategy = "WeightedRoundRobin"
	LeastConnections   Strategy = "LeastConnections"
)

func GetLoadBalancer(strategy Strategy, servers []Server) LoadBalancer {
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
