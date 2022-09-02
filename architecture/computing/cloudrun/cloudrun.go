package cloudrun

import (
	"context"
	"fmt"
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
	containers []Container
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

		revisions = append(revisions, Revision{containers: containers})
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
			cpuLimit := container.limits["cpu"]
			cpuLimitNum := cpuLimit[:len(cpuLimit)-1] // Drop "m" from the last
			cpuNum, err := strconv.Atoi(cpuLimitNum)
			if err != nil {
				return CloudRun{}, false
			}

			memLimit := container.limits["memory"]
			memLimitNum := memLimit[:len(memLimit)-2] // Drop Mi and Gi from the last
			memNum, err := strconv.Atoi(memLimitNum)
			if err != nil {
				return CloudRun{}, false
			}

			totalCost += float64(cpuNum)/1000*cost.SERVERLESS_COST_PER_CPU_CORE + float64(memNum)*cost.SERVERLESS_COST_PER_MEM_MIB
		}
	}

	return CloudRun{
		id:     path.Base(service.name),
		region: service.location,
		cost:   totalCost,
	}, true
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
	return ""
}

func (r CloudRun) GetZone() string {
	return ""
}
