package cloudfunctions

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path"
	"strconv"
	"strings"

	functions "cloud.google.com/go/functions/apiv2"
	"github.com/mittz/roleplay-webapp-assess/cost"
	"google.golang.org/api/iterator"
	functionspb "google.golang.org/genproto/googleapis/cloud/functions/v2"
)

var availableCPUs = map[string]float64{
	"128M": 0.083,
	"256M": 0.167,
	"512M": 0.333,
	"1G":   0.583,
	"2G":   1,
	"4G":   2,
	"8G":   2,
	"16G":  4,
}

type CloudFunctions struct {
	id     string
	region string
	cost   float64
}

type Function struct {
	name             string
	availableMemory  string
	maxInstanceCount int
	minInstanceCount int
}

type Revision struct {
	containers []Container
}

type Container struct {
	limits map[string]string
}

func getFunction(projectID string, hostName string) (Function, error) {
	ctx := context.Background()
	c, err := functions.NewFunctionClient(ctx)
	if err != nil {
		return Function{}, err
	}
	defer c.Close()

	req := &functionspb.ListFunctionsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/-", projectID),
	}
	it := c.ListFunctions(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return Function{}, err
		}

		u, err := url.Parse(resp.GetServiceConfig().GetUri())
		if err != nil {
			return Function{}, err
		}

		if u.Host == hostName {
			return Function{
				name:             resp.GetName(),
				availableMemory:  resp.ServiceConfig.GetAvailableMemory(),
				maxInstanceCount: int(resp.ServiceConfig.GetMaxInstanceCount()),
				minInstanceCount: int(resp.ServiceConfig.GetMinInstanceCount()),
			}, nil
		}
	}

	return Function{}, fmt.Errorf("Cloud Functions Function was not found.")
}

func GetCloudFunctions(projectID string, hostName string) (CloudFunctions, bool) {
	function, err := getFunction(projectID, hostName)
	if err != nil {
		return CloudFunctions{}, false
	}

	mem, err := strconv.Atoi(function.availableMemory[:len(function.availableMemory)-1]) // Drop "M", "G"
	if err != nil {
		return CloudFunctions{}, false
	}

	cpu, ok := availableCPUs[function.availableMemory]
	if !ok {
		log.Printf("Unknown memory spec")
		return CloudFunctions{}, false
	}

	avgInstanceCount := (function.maxInstanceCount + function.minInstanceCount) / 2

	return CloudFunctions{
		id:     path.Base(function.name),
		region: strings.Split(function.name, "/")[3],
		cost:   (cpu*cost.SERVERLESS_COST_PER_CPU_CORE + float64(mem)*cost.SERVERLESS_COST_PER_MEM_MIB) * float64(avgInstanceCount),
	}, true
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
