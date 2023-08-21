package main

import (
	"sync/atomic"
)

type leastConnectionsLoadBalancer struct {
	servers []server
}

func (s *server) increaseConnections() {
	atomic.AddInt64(&s.activeConnections, 1)
}

func (s *server) decreaseConnections() {
	atomic.AddInt64(&s.activeConnections, -1)
}

func (lb *leastConnectionsLoadBalancer) pickServer() *server {
	var best *server
	for i := 0; i < len(lb.servers); i++ {
		s := &lb.servers[i]
		if best == nil || s.activeConnections < best.activeConnections {
			best = s
		}
	}

	return best
}
