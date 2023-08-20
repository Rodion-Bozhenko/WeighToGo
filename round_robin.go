package main

type roundRobinLoadBalancer struct {
	servers []server
	count   int
}

func (rb *roundRobinLoadBalancer) pickServer() *server {
	index := rb.count % len(rb.servers)
	rb.count++
	best := &rb.servers[index]
	return best
}
