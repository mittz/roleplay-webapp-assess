package architecture

type AlloyDB struct {
	id   string
	cost float64
}

func GetAlloyDB(projectID string) (AlloyDB, bool) {
	db := AlloyDB{}
	return db, false
}

func (r AlloyDB) GetID() string {
	return r.id
}

func (r AlloyDB) GetAvailabilityRate() int {
	return 0
}

func (r AlloyDB) GetCost() float64 {
	return r.cost
}

func (r AlloyDB) SetCost(cost float64) {
	r.cost = cost
}
