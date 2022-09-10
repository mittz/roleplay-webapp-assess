package cloudrun

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path"
	"strconv"
	"strings"

	asset "cloud.google.com/go/asset/apiv1"
	run "cloud.google.com/go/run/apiv2"
	"github.com/mittz/roleplay-webapp-assess/cost"
	"google.golang.org/api/iterator"
	assetpb "google.golang.org/genproto/googleapis/cloud/asset/v1"
	runpb "google.golang.org/genproto/googleapis/cloud/run/v2"
)

type CloudRun struct {
	id     string
	region string
	cost   float64
}

type Service struct {
	name     string
	location string
}

type Revision struct {
	containers       []Container
	avgInstanceCount int
}

type Container struct {
	limits map[string]string
}

func getService(projectID string, hostName string) (Service, error) {
	scope := fmt.Sprintf("projects/%s", projectID)
	ctx := context.Background()

	client, err := asset.NewClient(ctx)
	if err != nil {
		return Service{}, err
	}
	defer client.Close()

	req := &assetpb.SearchAllResourcesRequest{
		Scope: scope,
		AssetTypes: []string{
			"run.googleapis.com/Service",
		},
	}

	it := client.SearchAllResources(ctx, req)
	for {
		resource, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return Service{}, err
		}

		u, err := url.Parse(resource.GetAdditionalAttributes().GetFields()["statusUrl"].GetStringValue())
		if err != nil {
			return Service{}, err
		}

		if u.Host == hostName && resource.GetLabels()["goog-managed-by"] != "cloudfunctions" {
			if err != nil {
				return Service{}, err
			}

			return Service{
				name:     strings.Join(strings.Split(resource.Name, "/")[3:], "/"), // Drop "//run.googleapis.com/"
				location: resource.GetLocation(),
			}, nil
		}
	}

	return Service{}, fmt.Errorf("Cloud Run Service was not found.")
}

func (s Service) GetRevisions() ([]Revision, error) {
	ctx := context.Background()
	c, err := run.NewRevisionsClient(ctx)
	if err != nil {
		return []Revision{}, err
	}
	defer c.Close()

	var revisions []Revision
	req := &runpb.ListRevisionsRequest{
		Parent: s.name,
	}
	it := c.ListRevisions(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []Revision{}, err
		}

		minInstanceCount := resp.GetScaling().GetMaxInstanceCount()
		maxInstanceCount := resp.GetScaling().GetMinInstanceCount()

		var containers []Container
		for _, condition := range resp.GetConditions() {
			if condition.GetType() == "ResourcesAvailable" && condition.GetState() == runpb.Condition_CONDITION_SUCCEEDED {
				for _, container := range resp.GetContainers() {
					containers = append(containers, Container{
						limits: container.Resources.GetLimits(),
					})
				}
			}
		}

		revisions = append(revisions, Revision{
			containers:       containers,
			avgInstanceCount: (int(minInstanceCount) + int(maxInstanceCount)) / 2,
		})
	}

	return revisions, nil
}

func GetCloudRun(projectID string, hostName string) (CloudRun, bool) {
	service, err := getService(projectID, hostName)
	if err != nil {
		return CloudRun{}, false
	}

	revisions, err := service.GetRevisions()
	if err != nil {
		return CloudRun{}, false
	}

	var totalCost float64
	for _, revision := range revisions {
		for _, container := range revision.containers {
			cpuLimit, memLimit := container.limits["cpu"], container.limits["memory"]

			var cpuNum, memNum int
			if strings.Contains(cpuLimit, "m") {
				cpuNum, err = strconv.Atoi(strings.TrimRight(cpuLimit, "m"))
				if err != nil {
					log.Printf("Error: %v", err)
					return CloudRun{}, false
				}
				cpuNum /= 1000 // 1000, 2000, 3000 -> 1, 2, 3
			} else {
				cpuNum, err = strconv.Atoi(cpuLimit)
				if err != nil {
					log.Printf("Error: %v", err)
					return CloudRun{}, false
				}
			}

			if strings.Contains(memLimit, "Mi") {
				memNum, err = strconv.Atoi(strings.TrimRight(memLimit, "Mi"))
				if err != nil {
					return CloudRun{}, false
				}
			} else if strings.Contains(memLimit, "Gi") {
				memNum, err = strconv.Atoi(strings.TrimRight(memLimit, "Gi"))
				if err != nil {
					return CloudRun{}, false
				}
				memNum *= 1024 // Gi to Mi
			} else {
				log.Printf("Unexpected case was found in memory limit: %s", memLimit)
				return CloudRun{}, false
			}

			totalCost += (float64(cpuNum)*cost.SERVERLESS_COST_PER_CPU_CORE + float64(memNum)*cost.SERVERLESS_COST_PER_MEM_MIB) * float64(revision.avgInstanceCount)
		}
	}

	return CloudRun{
		id:     path.Base(service.name),
		region: service.location,
		cost:   totalCost,
	}, true
}

func GetCloudRunService(projectID string, region string, name string) (CloudRun, error) {
	ctx := context.Background()
	c, err := run.NewServicesClient(ctx)
	if err != nil {
		return CloudRun{}, err
	}
	defer c.Close()

	req := &runpb.GetServiceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/services/%s", projectID, region, name),
	}
	resp, err := c.GetService(ctx, req)
	if err != nil {
		return CloudRun{}, nil
	}

	u, err := url.Parse(resp.GetUri())
	if err != nil {
		return CloudRun{}, err
	}

	x, exist := GetCloudRun(projectID, u.Host)
	if exist {
		return x, nil
	}

	return CloudRun{}, fmt.Errorf("Cloud Run Service: %s doesn't exist in %s - %s", name, projectID, region)
}

func (r CloudRun) GetID() string {
	return r.id
}

func (r CloudRun) SetCost(cost float64) {
	r.cost = cost
}

func (r CloudRun) GetCost() float64 {
	return r.cost
}

func (r CloudRun) GetRegion() string {
	return r.region
}

func (r CloudRun) GetZone() string {
	return ""
}
