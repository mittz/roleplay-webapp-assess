package cloudspanner

import (
	"context"
	"fmt"
	"strings"

	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"github.com/mittz/roleplay-webapp-assess/cost"
	"google.golang.org/api/iterator"
	instancepb "google.golang.org/genproto/googleapis/spanner/admin/instance/v1"
)

type CloudSpanner struct {
	id               string
	cost             float64
	availabilityRate int
}

type Instance struct {
	Name            string
	ProcessingUnits int32
	Config          string
}

func getInstances(projectID string) ([]Instance, error) {
	ctx := context.Background()
	c, err := instance.NewInstanceAdminClient(ctx)
	if err != nil {
		return []Instance{}, err
	}
	defer c.Close()

	req := &instancepb.ListInstancesRequest{
		Parent: fmt.Sprintf("projects/%s", projectID),
	}

	var instances []Instance
	it := c.ListInstances(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []Instance{}, err
		}

		instances = append(instances, Instance{
			Name:            resp.GetName(),
			ProcessingUnits: resp.GetProcessingUnits(),
			Config:          resp.GetConfig(),
		})
	}

	if len(instances) == 0 {
		return []Instance{}, fmt.Errorf("Cloud Spanner instance was not found.")
	}

	return instances, nil
}

func GetCloudSpanner(projectID string) (CloudSpanner, bool) {
	instances, err := getInstances(projectID)
	if err != nil {
		return CloudSpanner{}, false
	}

	instance := instances[0] // Pick up one if there are multiple instances

	totalCost := float64(instance.ProcessingUnits) * cost.SPANNER_COST_PER_PROCESSING_UNIT
	availabilityRate := 1
	if strings.Contains(instance.Config, "regional") {
		availabilityRate = 2
	} else {
		availabilityRate = 3
	}

	return CloudSpanner{
		id:               instance.Name,
		cost:             totalCost,
		availabilityRate: availabilityRate,
	}, true
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
