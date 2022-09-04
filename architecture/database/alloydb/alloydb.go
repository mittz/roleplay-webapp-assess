package alloydb

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/mittz/roleplay-webapp-assess/cost"
)

// CPU count: Memory MiB
var instanceTypes = map[int]int{
	2:  16 * 1024,
	4:  32 * 1024,
	8:  64 * 1024,
	16: 128 * 1024,
	32: 256 * 1024,
	64: 512 * 1024,
}

type AlloyDB struct {
	id               string
	cost             float64
	availabilityRate int
}

type Cluster struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

func getClusters(projectID string) ([]Cluster, error) {
	out, err := exec.Command(
		"gcloud",
		"beta",
		"alloydb",
		"clusters",
		"list",
		fmt.Sprintf("--project=%s", projectID),
		"--format=json",
	).Output()
	if err != nil {
		log.Println(err)
		return []Cluster{}, err
	}

	var clusters []Cluster
	if err := json.Unmarshal(out, &clusters); err != nil {
		return []Cluster{}, err
	}

	if len(clusters) == 0 {
		return []Cluster{}, fmt.Errorf("AlloyDB Cluster was not found.")
	}

	return clusters, nil
}

type Instance struct {
	InstanceType   string         `json:"instanceType"`
	MachineConfig  MachineConfig  `json:"machineConfig"`
	ReadPoolConfig ReadPoolConfig `json:"readPoolConfig"`
}

type MachineConfig struct {
	CPUCount int `json:"cpuCount"`
}

type ReadPoolConfig struct {
	NodeCount int `json:"nodeCount"`
}

func (c Cluster) GetInstances() ([]Instance, error) {
	// "projects/<projectID>/locations/<region>/clusters/<clusterName>"
	names := strings.Split(c.Name, "/")
	projectID, region, clusterName := names[1], names[3], names[5]
	out, err := exec.Command(
		"gcloud",
		"beta",
		"alloydb",
		"instances",
		"list",
		fmt.Sprintf("--cluster=%s", clusterName),
		fmt.Sprintf("--region=%s", region),
		fmt.Sprintf("--project=%s", projectID),
		"--format=json",
	).Output()
	if err != nil {
		return []Instance{}, err
	}

	var instances []Instance
	if err := json.Unmarshal(out, &instances); err != nil {
		return []Instance{}, err
	}

	return instances, nil
}

func GetAlloyDB(projectID string) (AlloyDB, bool) {
	clusters, err := getClusters(projectID)
	if err != nil {
		return AlloyDB{}, false
	}

	cluster := clusters[0] // Pick up one cluster if there are multiple ones

	instances, err := cluster.GetInstances()
	if err != nil {
		return AlloyDB{}, false
	}

	var totalCost float64
	for _, instance := range instances {
		cpuCount := instance.MachineConfig.CPUCount
		nodeCount := 2 // PRIMARY_INSTANCE with HA = 2 nodes
		if instance.InstanceType == "READ_POOL" {
			nodeCount = instance.ReadPoolConfig.NodeCount
		}

		totalCost += (float64(cpuCount)*cost.ALLOYDB_COST_PER_CPU_CORE + float64(instanceTypes[cpuCount])*cost.ALLOYDB_COST_PER_MEM_MIB) * float64(nodeCount)
	}

	return AlloyDB{id: clusters[0].UID, availabilityRate: 2, cost: totalCost}, true
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
