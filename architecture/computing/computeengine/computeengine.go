package computeengine

import (
	"context"
	"log"
	"path"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"github.com/mittz/roleplay-webapp-assess/cost"
	"github.com/mittz/roleplay-webapp-assess/utils"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type ComputeEngine struct {
	id     string
	region string
	zone   string
	cost   float64
}

type Resource struct {
	IsSharedCpu bool
	CPU         int
	MemoryMib   int
}

func calcCost(resource Resource) float64 {
	sharedRate := 1.0
	if resource.IsSharedCpu {
		sharedRate = 0.5
	}

	return (float64(resource.CPU)*cost.GCE_COST_PER_CPU_CORE + float64(resource.MemoryMib)*cost.GCE_COST_PER_MEM_MIB) * sharedRate
}

func getMachineType(projectID string, zone string, machineType string) (Resource, error) {
	ctx := context.Background()
	c, err := compute.NewMachineTypesRESTClient(ctx)
	if err != nil {
		return Resource{}, nil
	}
	defer c.Close()

	req := &computepb.GetMachineTypeRequest{
		Project:     projectID,
		MachineType: machineType,
		Zone:        zone,
	}
	resp, err := c.Get(ctx, req)
	if err != nil {
		return Resource{}, nil
	}

	return Resource{
		IsSharedCpu: resp.GetIsSharedCpu(),
		CPU:         int(resp.GetGuestCpus()),
		MemoryMib:   int(resp.GetMemoryMb()),
	}, nil
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
							resource, err := getMachineType(projectID, strings.Split(instance.GetMachineType(), "/")[8], path.Base(instance.GetMachineType()))
							if err != nil {
								log.Println(err)
								return ComputeEngine{}, false
							}

							return ComputeEngine{
								id:     instance.GetName(),
								zone:   path.Base(instance.GetZone()),
								region: utils.GetRegionFromZone(path.Base(instance.GetZone())),
								cost:   calcCost(resource),
							}, true
						}
					}
				}
			}
		}
	}

	return ComputeEngine{}, false
}

func GetComputeInstance(projectID string, zone string, name string) (ComputeEngine, error) {
	ctx := context.Background()
	c, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return ComputeEngine{}, err
	}
	defer c.Close()

	req := &computepb.GetInstanceRequest{
		Project:  projectID,
		Zone:     zone,
		Instance: name,
	}
	resp, err := c.Get(ctx, req)
	if err != nil {
		return ComputeEngine{}, err
	}

	return ComputeEngine{
		id:     resp.GetName(),
		region: utils.GetRegionFromZone(path.Base(resp.GetZone())),
		zone:   path.Base(resp.GetZone()),
		cost:   0,
	}, nil
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
