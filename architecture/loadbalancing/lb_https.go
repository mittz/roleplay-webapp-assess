package loadbalancing

import (
	"context"
	"fmt"
	"path"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	computing "github.com/mittz/roleplay-webapp-assess/architecture/computing"
	"github.com/mittz/roleplay-webapp-assess/architecture/computing/computeengine"
	"google.golang.org/api/iterator"
	computepb "google.golang.org/genproto/googleapis/cloud/compute/v1"
)

type LoadBalancingHTTPS struct {
	id       string
	backends []computing.Computing
}

type ForwardingRule struct {
	Name       string
	Region     string
	TargetPool string
}

func GetLoadBalancingHTTPS(projectID string, hostIP string) (LoadBalancingHTTPS, bool) {
	forwardingRule, err := getForwardingRules(projectID, hostIP)
	if err != nil {
		return LoadBalancingHTTPS{}, false
	}

	lb := LoadBalancingHTTPS{
		id:       forwardingRule.Name,
		backends: getRegionBackendServices(projectID, forwardingRule),
	}

	return lb, true
}

func getForwardingRules(projectID string, hostIP string) (*ForwardingRule, error) {
	ctx := context.Background()
	c, err := compute.NewForwardingRulesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	req := &computepb.AggregatedListForwardingRulesRequest{
		Project: projectID,
	}
	it := c.AggregatedList(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, rule := range resp.Value.GetForwardingRules() {
			if rule.GetIPAddress() == hostIP {
				return &ForwardingRule{
					Name:       rule.GetName(),
					Region:     path.Base(rule.GetRegion()),
					TargetPool: path.Base(rule.GetTarget()),
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("None forwarding rule matched to the host ipaddress")
}

func getRegionBackendServices(projectID string, forwardingRule *ForwardingRule) []computing.Computing {
	ctx := context.Background()
	c, err := compute.NewTargetPoolsRESTClient(ctx)
	if err != nil {
		return []computing.Computing{}
	}
	defer c.Close()

	req := &computepb.AggregatedListTargetPoolsRequest{
		Project: projectID,
	}

	var backends []computing.Computing
	it := c.AggregatedList(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return []computing.Computing{}
		}

		for _, pool := range resp.Value.TargetPools {
			if pool.GetName() == forwardingRule.TargetPool {
				for _, instance := range pool.GetInstances() {
					name := path.Base(instance)
					paths := strings.Split(instance, "/")
					// https://www.googleapis.com/compute/v1/projects/<Project Name>/zones/<Zone>/instances/<Instance Name>
					zone := paths[8]
					i, err := computeengine.GetComputeInstance(projectID, zone, name)
					if err != nil {
						return []computing.Computing{}
					}

					backends = append(backends, i)
				}
			}
		}
	}

	return backends
}

func (r LoadBalancingHTTPS) GetID() string {
	return r.id
}

func (r LoadBalancingHTTPS) GetBackends() []computing.Computing {
	return r.backends
}
