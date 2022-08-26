package architecture

type AppEngine struct {
	id   string
	cost float64
}

func GetAppEngine(projectID string, hostIP string) (AppEngine, bool) {
	return AppEngine{}, false
}

func (r AppEngine) GetID() string {
	return r.id
}

func (r AppEngine) SetCost(cost float64) {
	r.cost = cost
}

func (r AppEngine) GetCost() float64 {
	return r.cost
}

func (r AppEngine) GetRegion() string {
	return ""
}

func (r AppEngine) GetZone() string {
	return ""
}
