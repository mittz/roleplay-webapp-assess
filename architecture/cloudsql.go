package architecture

import (
	"context"
	"log"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/sqladmin/v1"
)

type CloudSQL struct {
	id               string
	cost             float64
	availabilityRate int
}

func GetCloudSQL(projectID string) (CloudSQL, bool) {
	ctx := context.Background()

	// Create an http.Client that uses Application Default Credentials.
	hc, err := google.DefaultClient(ctx, sqladmin.SqlserviceAdminScope)
	if err != nil {
		log.Println(err)
		return CloudSQL{}, false
	}

	// Create the Google Cloud SQL service.
	service, err := sqladmin.New(hc)
	if err != nil {
		log.Println(err)
		return CloudSQL{}, false
	}

	// List instances for the project ID.
	instances, err := service.Instances.List(projectID).Do()
	if err != nil {
		log.Println(err)
		return CloudSQL{}, false
	}

	if len(instances.Items) == 0 {
		return CloudSQL{}, false
	}

	var primaryInstance *sqladmin.DatabaseInstance
	replicaInstances := make(map[string]*sqladmin.DatabaseInstance)
	for _, instance := range instances.Items {
		if instance.State == "RUNNABLE" && instance.InstanceType == "CLOUD_SQL_INSTANCE" {
			primaryInstance = instance
			break
		}
	}

	for _, instance := range instances.Items {
		if instance.State == "RUNNABLE" && instance.InstanceType == "READ_REPLICA_INSTANCE" && instance.MasterInstanceName == primaryInstance.Name {
			replicaInstances[instance.Name] = instance
		}
	}

	if primaryInstance == nil {
		return CloudSQL{}, false
	}

	for _, replica := range replicaInstances {
		if replica.Region != primaryInstance.Region {
			return CloudSQL{
				id:               primaryInstance.Name,
				cost:             0,
				availabilityRate: 3,
			}, true
		}
	}

	if primaryInstance.Settings.AvailabilityType == "REGIONAL" {
		return CloudSQL{
			id:               primaryInstance.Name,
			cost:             0,
			availabilityRate: 2,
		}, true
	}

	return CloudSQL{
		id:               primaryInstance.Name,
		cost:             0,
		availabilityRate: 1,
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
