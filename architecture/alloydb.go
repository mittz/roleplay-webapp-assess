package architecture

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

type AlloyDB struct {
	id               string
	cost             float64
	availabilityRate int
}

type Cluster struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

func GetAlloyDB(projectID string) (AlloyDB, bool) {
	outCluster, err := exec.Command(
		"gcloud",
		"beta",
		"alloydb",
		"clusters",
		"list",
		fmt.Sprintf("--project=%s", projectID),
		"--format=json",
	).Output()
	if err != nil {
		log.Printf("GetAlloyDB: %v", err)
		return AlloyDB{}, false
	}

	var clusters []Cluster
	if err := json.Unmarshal(outCluster, &clusters); err != nil {
		log.Println(err)
		return AlloyDB{}, false
	}

	if len(clusters) == 0 {
		return AlloyDB{}, false
	}

	return AlloyDB{id: clusters[0].UID, availabilityRate: 2}, true
}

func (r AlloyDB) GetID() string {
	return r.id
}

func (r AlloyDB) GetAvailabilityRate() int {
	return r.availabilityRate
}

func (r AlloyDB) GetCost() float64 {
	return r.cost
}

func (r AlloyDB) SetCost(cost float64) {
	r.cost = cost
}
