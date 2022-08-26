package architecture

type LoadBalancingHTTPS struct {
	id string
}

func GetLoadBalancingHTTPS(projectID string, hostIP string) (LoadBalancingHTTPS, bool) {
	return LoadBalancingHTTPS{}, false
}

func (r LoadBalancingHTTPS) GetID() string {
	return r.id
}

func (r LoadBalancingHTTPS) GetBackends() []Computing {
	return nil
}
