package appengine

import (
	"context"
	"fmt"

	appengine "cloud.google.com/go/appengine/apiv1"
	"github.com/mittz/roleplay-webapp-assess/cost"
	"google.golang.org/api/iterator"
	appenginepb "google.golang.org/genproto/googleapis/appengine/v1"
)

type AppEngine struct {
	id     string
	region string
	cost   float64
}

type Application struct {
	name   string
	region string
}

type Service struct {
	name string
}

type Version struct {
	name             string
	instanceClass    string
	avgInstanceCount int
}

// https://cloud.google.com/appengine/docs/standard#instance_classes
var costTables = map[string]float64{
	"F1":    cost.SERVERLESS_COST_PER_CPU_CORE*0.6 + cost.SERVERLESS_COST_PER_MEM_MIB*256,  // CPU: 600 MHz Mem: 256 MB
	"F2":    cost.SERVERLESS_COST_PER_CPU_CORE*1.2 + cost.SERVERLESS_COST_PER_MEM_MIB*512,  // CPU: 1.2 GHz Mem: 512 MB
	"F4":    cost.SERVERLESS_COST_PER_CPU_CORE*2.4 + cost.SERVERLESS_COST_PER_MEM_MIB*1024, // CPU: 2.4 GHz Mem: 1024 MB
	"F4_1G": cost.SERVERLESS_COST_PER_CPU_CORE*2.4 + cost.SERVERLESS_COST_PER_MEM_MIB*2048, // CPU: 2.4 GHz Mem: 2048 MB
	"B1":    cost.SERVERLESS_COST_PER_CPU_CORE*0.6 + cost.SERVERLESS_COST_PER_MEM_MIB*256,  // CPU: 600 MHz Mem: 256 MB
	"B2":    cost.SERVERLESS_COST_PER_CPU_CORE*1.2 + cost.SERVERLESS_COST_PER_MEM_MIB*512,  // CPU: 600 MHz Mem: 256 MB
	"B4":    cost.SERVERLESS_COST_PER_CPU_CORE*2.4 + cost.SERVERLESS_COST_PER_MEM_MIB*1024, // CPU: 2.4 GHz Mem: 1024 MB
	"B4_1G": cost.SERVERLESS_COST_PER_CPU_CORE*2.4 + cost.SERVERLESS_COST_PER_MEM_MIB*2048, // CPU: 2.4 GHz Mem: 2048 MB
	"B8":    cost.SERVERLESS_COST_PER_CPU_CORE*4.8 + cost.SERVERLESS_COST_PER_MEM_MIB*2048, // CPU: 4.8 GHz Mem: 2048 MB
}

func getApplication(projectID string, hostName string) (Application, error) {
	ctx := context.Background()
	c, err := appengine.NewApplicationsClient(ctx)
	if err != nil {
		return Application{}, err
	}
	defer c.Close()

	req := &appenginepb.GetApplicationRequest{
		Name: fmt.Sprintf("apps/%s", projectID),
	}
	resp, err := c.GetApplication(ctx, req)
	if err != nil {
		return Application{}, err
	}

	if resp.GetDefaultHostname() != hostName {
		return Application{}, fmt.Errorf("App Engine Application was not found.")
	}

	var location string
	switch v := resp.LocationId; v {
	case "europe-west":
		location = "europe-west1"
	case "us-central":
		location = "us-central1"
	default:
		location = v
	}

	return Application{
		name:   resp.GetName(),
		region: location,
	}, nil
}

func (a Application) GetServices() ([]Service, error) {
	ctx := context.Background()
	c, err := appengine.NewServicesClient(ctx)
	if err != nil {
		return []Service{}, err
	}
	defer c.Close()

	req := &appenginepb.ListServicesRequest{
		Parent: a.name,
	}

	var services []Service
	it := c.ListServices(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []Service{}, err
		}

		services = append(services, Service{name: resp.GetName()})
	}

	return services, nil
}

func (s Service) GetVersions() ([]Version, error) {
	ctx := context.Background()
	c, err := appengine.NewVersionsClient(ctx)
	if err != nil {
		return []Version{}, err
	}
	defer c.Close()

	req := &appenginepb.ListVersionsRequest{
		Parent: s.name,
	}

	var versions []Version
	it := c.ListVersions(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []Version{}, err
		}

		minInstanceCount := resp.GetAutomaticScaling().GetMinTotalInstances()
		maxInstanceCount := resp.GetAutomaticScaling().GetMaxTotalInstances()
		if maxInstanceCount == 0 {
			maxInstanceCount = 100 // Default value which is the same as one of Cloud Run
		}

		versions = append(versions, Version{
			name:             resp.GetName(),
			instanceClass:    resp.GetInstanceClass(),
			avgInstanceCount: (int(minInstanceCount) + int(maxInstanceCount)) / 2,
		})
	}

	return versions, nil
}

func GetAppEngine(projectID string, hostName string) (AppEngine, bool) {
	application, err := getApplication(projectID, hostName)
	if err != nil {
		return AppEngine{}, false
	}

	services, err := application.GetServices()
	if err != nil {
		return AppEngine{}, false
	}

	var versions []Version
	for _, service := range services {
		vs, err := service.GetVersions()
		if err != nil {
			return AppEngine{}, false
		}

		versions = append(versions, vs...)
	}

	var cost float64
	for _, version := range versions {
		cost += costTables[version.instanceClass] * float64(version.avgInstanceCount)
	}

	return AppEngine{
		id:     application.name,
		region: application.region,
		cost:   cost,
	}, true
}

func (r AppEngine) GetID() string {
	return r.id
}

func (r AppEngine) SetCost(cost float64) {
	r.cost = cost
}

func (r AppEngine) GetCost() float64 {
	return r.cost
}

func (r AppEngine) GetRegion() string {
	return ""
}

func (r AppEngine) GetZone() string {
	return ""
}
