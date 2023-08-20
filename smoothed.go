package main

type smoothedLoadBalancer struct {
	servers []server
}

func (sw *smoothedLoadBalancer) pickServer() *server {
	var best *server
	total := 0
	for i := 0; i < len(sw.servers); i++ {
		total += sw.servers[i].currentWeight
		sw.servers[i].currentWeight += sw.servers[i].weight

		if best == nil || sw.servers[i].currentWeight > best.currentWeight {
			best = &sw.servers[i]
		}
	}
	best.currentWeight -= total
	return best
}
