package loadbalancer

import (
	"sync/atomic"
)

type LeastConnectionsLoadBalancer struct {
	Servers []Server
}

func (s *Server) IncreaseConnections() {
	atomic.AddInt64(&s.ActiveConnections, 1)
}

func (s *Server) DecreaseConnections() {
	atomic.AddInt64(&s.ActiveConnections, -1)
}

func (lb *LeastConnectionsLoadBalancer) PickServer() *Server {
	var best *Server
	for i := 0; i < len(lb.Servers); i++ {
		s := &lb.Servers[i]
		if best == nil || s.ActiveConnections < best.ActiveConnections {
			best = s
		}
	}

	return best
}
