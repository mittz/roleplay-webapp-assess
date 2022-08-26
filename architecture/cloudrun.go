package architecture

type CloudRun struct {
	id   string
	cost float64
}

func GetCloudRun(projectID string, hostIP string) (CloudRun, bool) {
	return CloudRun{}, false
}

func (r CloudRun) GetID() string {
	return r.id
}

func (r CloudRun) SetCost(cost float64) {
	r.cost = cost
}

func (r CloudRun) GetCost() float64 {
	return r.cost
}

func (r CloudRun) GetRegion() string {
	return ""
}

func (r CloudRun) GetZone() string {
	return ""
}
