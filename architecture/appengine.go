package architecture

import (
	"context"
	"fmt"
	"log"

	appengine "cloud.google.com/go/appengine/apiv1"
	appenginepb "google.golang.org/genproto/googleapis/appengine/v1"
)

type AppEngine struct {
	id     string
	region string
	cost   float64
}

func GetAppEngine(projectID string, hostName string) (AppEngine, bool) {
	ctx := context.Background()
	c, err := appengine.NewApplicationsClient(ctx)
	if err != nil {
		log.Println(err)
		return AppEngine{}, false
	}
	defer c.Close()

	req := &appenginepb.GetApplicationRequest{
		Name: fmt.Sprintf("apps/%s", projectID),
	}
	resp, err := c.GetApplication(ctx, req)
	if err != nil {
		log.Println(err)
		return AppEngine{}, false
	}

	if resp.GetDefaultHostname() == hostName {
		var location string
		switch v := resp.LocationId; v {
		case "europe-west":
			location = "europe-west1"
		case "us-central":
			location = "us-central1"
		default:
			location = v
		}

		return AppEngine{
			id:     resp.GetId(),
			region: location,
			cost:   0,
		}, true
	}

	return AppEngine{}, false
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
