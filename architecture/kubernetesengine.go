package architecture

type KubernetesEngine struct {
	id   string
	cost float64
}

func GetKubernetesEngine(projectID string, hostIP string) (KubernetesEngine, bool) {
	return KubernetesEngine{}, false
}

func (r KubernetesEngine) GetID() string {
	return r.id
}

func (r KubernetesEngine) SetCost(cost float64) {
	r.cost = cost
}

func (r KubernetesEngine) GetCost() float64 {
	return r.cost
}

func (r KubernetesEngine) GetRegion() string {
	return ""
}

func (r KubernetesEngine) GetZone() string {
	return ""
}
