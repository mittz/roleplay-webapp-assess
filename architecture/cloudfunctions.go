package architecture

type CloudFunctions struct {
	id   string
	cost float64
}

func GetCloudFunctions(projectID string, hostIP string) (CloudFunctions, bool) {
	return CloudFunctions{}, false
}

func (r CloudFunctions) GetID() string {
	return r.id
}

func (r CloudFunctions) SetCost(cost float64) {
	r.cost = cost
}

func (r CloudFunctions) GetCost() float64 {
	return r.cost
}

func (r CloudFunctions) GetRegion() string {
	return ""
}

func (r CloudFunctions) GetZone() string {
	return ""
}
