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

type CloudFunctions struct {
	id     string
	region string
	cost   float64
}

func GetCloudFunctions(projectID string, hostName string) (CloudFunctions, bool) {
	scope := fmt.Sprintf("projects/%s", projectID)
	ctx := context.Background()
	client, err := asset.NewClient(ctx)
	if err != nil {
		return CloudFunctions{}, false
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
			return CloudFunctions{}, false
		}

		u, err := url.Parse(resource.GetAdditionalAttributes().GetFields()["statusUrl"].GetStringValue())
		if err != nil {
			log.Println(err)
			return CloudFunctions{}, false
		}

		if u.Host == hostName && resource.GetLabels()["goog-managed-by"] == "cloudfunctions" {
			return CloudFunctions{
				id:     path.Base(resource.GetName()),
				region: resource.GetLocation(),
				cost:   0,
			}, true
		}
	}

	return CloudFunctions{}, false
}

func (r CloudFunctions) GetID() string {
	return r.id
}

func (r CloudFunctions) SetCost(cost float64) {
	r.cost = cost
}

func (r CloudFunctions) GetCost() float64 {
	return r.cost
}

func (r CloudFunctions) GetRegion() string {
	return ""
}

func (r CloudFunctions) GetZone() string {
	return ""
}
