package cloudsql

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mittz/roleplay-webapp-assess/cost"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sqladmin/v1"
)

type CloudSQL struct {
	id               string
	cost             float64
	availabilityRate int
}

func getPrimaryInstance(projectID string) (*sqladmin.DatabaseInstance, error) {
	ctx := context.Background()

	// Create an http.Client that uses Application Default Credentials.
	hc, err := google.DefaultClient(ctx, sqladmin.SqlserviceAdminScope)
	if err != nil {
		return nil, err
	}

	// Create the Google Cloud SQL service.
	service, err := sqladmin.New(hc)
	if err != nil {
		return nil, err
	}

	// List instances for the project ID.
	instances, err := service.Instances.List(projectID).Do()
	if err != nil {
		return nil, err
	}

	if instances == nil {
		return nil, fmt.Errorf("Cloud SQL Instance was not found.")
	}

	for _, instance := range instances.Items {
		if instance.State == "RUNNABLE" && instance.InstanceType == "CLOUD_SQL_INSTANCE" {
			return instance, nil
		}
	}

	return nil, fmt.Errorf("Cloud SQL Instance was not found.")
}

func getReplicaInstances(projectID string, primaryInstanceName string) ([]*sqladmin.DatabaseInstance, error) {
	ctx := context.Background()

	// Create an http.Client that uses Application Default Credentials.
	hc, err := google.DefaultClient(ctx, sqladmin.SqlserviceAdminScope)
	if err != nil {
		return nil, err
	}

	// Create the Google Cloud SQL service.
	service, err := sqladmin.New(hc)
	if err != nil {
		return nil, err
	}

	// List instances for the project ID.
	instances, err := service.Instances.List(projectID).Do()
	if err != nil {
		return nil, err
	}

	if instances == nil {
		return nil, fmt.Errorf("Cloud SQL Instance was not found.")
	}

	var replicaInstances []*sqladmin.DatabaseInstance
	for _, instance := range instances.Items {
		if instance.State == "RUNNABLE" && instance.InstanceType == "READ_REPLICA_INSTANCE" && instance.MasterInstanceName == primaryInstanceName {
			replicaInstances = append(replicaInstances, instance)
		}
	}

	return replicaInstances, nil
}

func GetCloudSQL(projectID string) (CloudSQL, bool) {
	primaryInstance, err := getPrimaryInstance(projectID)
	if err != nil {
		return CloudSQL{}, false
	}

	replicaInstances, err := getReplicaInstances(projectID, primaryInstance.Name)
	if err != nil {
		return CloudSQL{}, false
	}

	haRate := 1
	if primaryInstance.FailoverReplica != nil && primaryInstance.FailoverReplica.Available {
		haRate = 2
	}

	primaryInstaceTier := strings.Split(primaryInstance.Settings.Tier, "-")

	cpu, err := strconv.Atoi(primaryInstaceTier[2])
	if err != nil {
		return CloudSQL{}, false
	}

	mem, err := strconv.Atoi(primaryInstaceTier[3])
	if err != nil {
		return CloudSQL{}, false
	}

	regions := map[string]interface{}{
		primaryInstance.Region: struct{}{},
	}

	totalCost := (float64(cpu)*cost.CLOUDSQL_COST_PER_CPU_CORE + float64(mem)*cost.CLOUDSQL_COST_PER_MEM_MIB) * float64(haRate)

	for _, replicaInstance := range replicaInstances {
		regions[replicaInstance.Region] = struct{}{}
		replicaInstanceTier := strings.Split(replicaInstance.Settings.Tier, "-")
		c, err := strconv.Atoi(replicaInstanceTier[2])
		if err != nil {
			return CloudSQL{}, false
		}

		m, err := strconv.Atoi(replicaInstanceTier[3])
		if err != nil {
			return CloudSQL{}, false
		}

		totalCost += float64(c)*cost.CLOUDSQL_COST_PER_CPU_CORE + float64(m)*cost.CLOUDSQL_COST_PER_MEM_MIB
	}

	var availabilityRate int
	if len(regions) > 1 {
		availabilityRate = 3
	} else if primaryInstance.Settings.AvailabilityType == "REGIONAL" {
		availabilityRate = 2
	} else {
		availabilityRate = 1
	}

	return CloudSQL{
		id:               primaryInstance.Name,
		cost:             totalCost,
		availabilityRate: availabilityRate,
	}, true
}

func (r CloudSQL) GetID() string {
	return r.id
}

func (r CloudSQL) GetAvailabilityRate() int {
	return r.availabilityRate
}

func (r CloudSQL) GetCost() float64 {
	return r.cost
}

func (r CloudSQL) SetCost(cost float64) {
	r.cost = cost
}
