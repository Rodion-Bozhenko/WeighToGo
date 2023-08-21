package loadbalancer

type RoundRobinLoadBalancer struct {
	Servers []Server
	Count   int
}

func (rb *RoundRobinLoadBalancer) PickServer() *Server {
	index := rb.Count % len(rb.Servers)
	rb.Count++
	best := &rb.Servers[index]
	return best
}
