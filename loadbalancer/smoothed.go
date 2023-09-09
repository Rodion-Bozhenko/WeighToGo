package loadbalancer

type SmoothedLoadBalancer struct {
	Servers []*Server
}

func (sw *SmoothedLoadBalancer) PickServer() *Server {
	var best *Server
	total := 0
	for i := 0; i < len(sw.Servers); i++ {
		total += sw.Servers[i].CurrentWeight
		sw.Servers[i].CurrentWeight += sw.Servers[i].Weight

		if best == nil || sw.Servers[i].CurrentWeight > best.CurrentWeight {
			best = sw.Servers[i]
		}
	}
	best.CurrentWeight -= total
	return best
}
