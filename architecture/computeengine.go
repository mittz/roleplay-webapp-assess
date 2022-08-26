package architecture

import (
	"context"
	"log"
	"path"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type ComputeEngine struct {
	id     string
	region string
	zone   string
	cost   float64
}

func GetComputeEngine(projectID string, hostIP string) (ComputeEngine, bool) {
	ctx := context.Background()
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		log.Println(err)
		return ComputeEngine{}, false
	}
	defer c.Close()

	req := &computepb.AggregatedListInstancesRequest{Project: projectID}
	it := c.AggregatedList(ctx, req)

	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Println(err)
			return ComputeEngine{}, false
		}

		for _, instance := range resp.Value.Instances {
			if instance.GetStatus() == "RUNNING" {
				for _, network := range instance.GetNetworkInterfaces() {
					for _, config := range network.GetAccessConfigs() {
						if config.GetNatIP() == hostIP {
							return ComputeEngine{
								id:     instance.GetFingerprint(),
								zone:   path.Base(instance.GetZone()),
								region: strings.Join(strings.Split(path.Base(instance.GetZone()), "-")[0:2], "-"),
								cost:   0,
								// For a custom machine type: zones/zone/machineTypes/custom-CPUS-MEMOR e.g. zones/us-central1-f/machineTypes/custom-4-5120
								// computeEngine.machineType = path.Base(instance.GetMachineType())
							}, true
						}
					}
				}
			}
		}
	}

	return ComputeEngine{}, false
}

func (r ComputeEngine) GetID() string {
	return r.id
}

func (r ComputeEngine) SetCost(cost float64) {
	r.cost = cost
}

func (r ComputeEngine) GetCost() float64 {
	return r.cost
}

func (r ComputeEngine) GetRegion() string {
	return r.region
}

func (r ComputeEngine) GetZone() string {
	return r.zone
}
