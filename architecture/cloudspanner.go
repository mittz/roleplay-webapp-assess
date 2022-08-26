package architecture

import (
	"context"
	"fmt"
	"log"
	"strings"

	asset "cloud.google.com/go/asset/apiv1"
	"google.golang.org/api/iterator"
	assetpb "google.golang.org/genproto/googleapis/cloud/asset/v1"
)

type CloudSpanner struct {
	id               string
	cost             float64
	availabilityRate int
}

func GetCloudSpanner(projectID string) (CloudSpanner, bool) {
	scope := fmt.Sprintf("projects/%s", projectID)
	ctx := context.Background()
	client, err := asset.NewClient(ctx)
	if err != nil {
		return CloudSpanner{}, false
	}
	defer client.Close()

	req := &assetpb.SearchAllResourcesRequest{
		Scope: scope,
		AssetTypes: []string{
			"spanner.googleapis.com/Instance",
		},
	}

	it := client.SearchAllResources(ctx, req)
	for {
		resource, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			log.Println(err)
			return CloudSpanner{}, false
		}

		location := resource.GetLocation()
		// TODO: Double check the condition below
		if !strings.Contains(location, "regional") {
			return CloudSpanner{
				id:               resource.GetName(),
				cost:             0,
				availabilityRate: 3,
			}, true
		} else {
			return CloudSpanner{
				id:               resource.GetName(),
				cost:             0,
				availabilityRate: 2,
			}, true
		}
	}

	return CloudSpanner{}, false
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
