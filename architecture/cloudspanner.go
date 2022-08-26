package architecture

type CloudSpanner struct {
	id   string
	cost float64
}

func GetCloudSpanner(projectID string) (CloudSpanner, bool) {
	db := CloudSpanner{}
	return db, false
}

func (r CloudSpanner) GetID() string {
	return r.id
}

func (r CloudSpanner) GetAvailabilityRate() int {
	return 0
}

func (r CloudSpanner) GetCost() float64 {
	return r.cost
}

func (r CloudSpanner) SetCost(cost float64) {
	r.cost = cost
}
