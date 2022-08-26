package architecture

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path"

	asset "cloud.google.com/go/asset/apiv1"
	"google.golang.org/api/iterator"
	assetpb "google.golang.org/genproto/googleapis/cloud/asset/v1"
)

type CloudRun struct {
	id     string
	region string
	cost   float64
}

func GetCloudRun(projectID string, hostName string) (CloudRun, bool) {
	scope := fmt.Sprintf("projects/%s", projectID)
	ctx := context.Background()
	client, err := asset.NewClient(ctx)
	if err != nil {
		log.Println(err)
		return CloudRun{}, false
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
			log.Println(err)
			return CloudRun{}, false
		}

		u, err := url.Parse(resource.GetAdditionalAttributes().GetFields()["statusUrl"].GetStringValue())
		if err != nil {
			log.Println(err)
			return CloudRun{}, false
		}

		if u.Host == hostName && resource.GetLabels()["goog-managed-by"] != "cloudfunctions" {
			return CloudRun{
				id:     path.Base(resource.GetName()),
				region: resource.GetLocation(),
				cost:   0,
			}, true
		}
	}

	return CloudRun{}, false
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
