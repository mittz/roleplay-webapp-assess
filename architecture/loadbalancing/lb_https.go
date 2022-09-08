package loadbalancing

import (
	"context"
	"fmt"
	"log"
	"path"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	computing "github.com/mittz/roleplay-webapp-assess/architecture/computing"
	"github.com/mittz/roleplay-webapp-assess/architecture/computing/cloudrun"
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

type TargetHTTPProxy struct {
	Name   string
	URLMap string
}

type URLMap struct {
	Name           string
	DefaultService string
}

type BackendService struct {
	Name     string
	Backends []*computepb.Backend
}

type Instance struct {
	Name   string
	Zone   string
	Status string
}

type Serverless struct {
	Name    string
	Region  string
	Service string
}

func GetLoadBalancingHTTPS(projectID string, hostIP string) (LoadBalancingHTTPS, bool) {
	forwardingRule, err := getForwardingRule(projectID, hostIP)
	if err != nil {
		log.Printf("GetLoadBalancingHTTPS - getForwardingRule: %v", err)
		return LoadBalancingHTTPS{}, false
	}

	targetHTTPProxy, err := forwardingRule.GetTargetHttpProxy(projectID)
	if err != nil {
		log.Printf("GetLoadBalancingHTTPS - GetTargetHttpProxy: %v", err)
		return LoadBalancingHTTPS{}, false
	}

	urlMap, err := targetHTTPProxy.GetURLMap(projectID)
	if err != nil {
		log.Printf("GetLoadBalancingHTTPS - GetURLMap: %v", err)
		return LoadBalancingHTTPS{}, false
	}

	backendService, err := urlMap.GetBackendService(projectID)
	if err != nil {
		log.Printf("GetLoadBalancingHTTPS - GetBackendService: %v", err)
		return LoadBalancingHTTPS{}, false
	}

	instances, err := backendService.ListInstances(projectID)
	if err != nil {
		log.Printf("GetLoadBalancingHTTPS - backendService.ListInstances: %v", err)
		return LoadBalancingHTTPS{}, false
	}

	serverlesses, err := backendService.ListServerlesses(projectID)
	if err != nil {
		log.Printf("GetLoadBalancingHTTPS - backendService.ListServerlesses: %v", err)
		return LoadBalancingHTTPS{}, false
	}

	var backends []computing.Computing
	for _, x := range instances {
		b, err := x.GetComputeEngine(projectID)
		if err != nil {
			log.Printf("GetLoadBalancingHTTPS - x.GetComputeEngine: %v", err)
			return LoadBalancingHTTPS{}, false
		}

		backends = append(backends, b)
	}

	for _, serverless := range serverlesses {
		b, err := serverless.Get(projectID)
		if err != nil {
			log.Printf("GetLoadBalancingHTTPS - serverless.Get: %v", err)
			return LoadBalancingHTTPS{}, false
		}

		backends = append(backends, b)
	}

	return LoadBalancingHTTPS{
		id:       forwardingRule.Name,
		backends: backends,
	}, true
}

func getForwardingRule(projectID string, hostIP string) (*ForwardingRule, error) {
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

func (f *ForwardingRule) GetTargetHttpProxy(projectID string) (*TargetHTTPProxy, error) {
	ctx := context.Background()
	c, err := compute.NewTargetHttpProxiesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	req := &computepb.GetTargetHttpProxyRequest{
		Project:         projectID,
		TargetHttpProxy: f.TargetPool,
	}
	resp, err := c.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	return &TargetHTTPProxy{
		Name:   resp.GetName(),
		URLMap: path.Base(resp.GetUrlMap()),
	}, nil
}

func (t *TargetHTTPProxy) GetURLMap(projectID string) (*URLMap, error) {
	ctx := context.Background()
	c, err := compute.NewUrlMapsRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	req := &computepb.GetUrlMapRequest{
		Project: projectID,
		UrlMap:  t.URLMap,
	}
	resp, err := c.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	return &URLMap{
		Name:           resp.GetName(),
		DefaultService: path.Base(resp.GetDefaultService()),
	}, nil
}

func (u *URLMap) GetBackendService(projectID string) (*BackendService, error) {
	ctx := context.Background()
	c, err := compute.NewBackendServicesRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	req := &computepb.GetBackendServiceRequest{
		Project:        projectID,
		BackendService: u.DefaultService,
	}
	resp, err := c.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	return &BackendService{
		Name:     resp.GetName(),
		Backends: resp.GetBackends(),
	}, nil
}

func (b *BackendService) ListInstances(projectID string) ([]*Instance, error) {
	var instances []*Instance
	for _, backend := range b.Backends {
		name := path.Base(backend.GetGroup())
		locationType := strings.Split(backend.GetGroup(), "/")[7]
		location := strings.Split(backend.GetGroup(), "/")[8]
		groupType := strings.Split(backend.GetGroup(), "/")[9]

		if groupType != "instanceGroups" {
			continue
		}

		if locationType == "regions" {
			x, err := getRegionInstanceGroup(projectID, name, location)
			if err != nil {
				return nil, err
			}

			instances = append(instances, x...)
		}

		if locationType == "zones" {
			x, err := getZoneInstanceGroup(projectID, name, location)
			if err != nil {
				return nil, err
			}

			instances = append(instances, x...)
		}
	}

	return instances, nil
}

func getRegionInstanceGroup(projectID string, name string, region string) ([]*Instance, error) {
	ctx := context.Background()
	c, err := compute.NewRegionInstanceGroupsRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	req := &computepb.ListInstancesRegionInstanceGroupsRequest{
		Project:       projectID,
		InstanceGroup: name,
		Region:        region,
	}
	it := c.ListInstances(ctx, req)
	var instances []*Instance
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		instances = append(instances, &Instance{
			Name:   path.Base(resp.GetInstance()),
			Zone:   strings.Split(resp.GetInstance(), "/")[8],
			Status: resp.GetStatus(),
		})
	}

	return instances, nil
}

func getZoneInstanceGroup(projectID string, name string, zone string) ([]*Instance, error) {
	ctx := context.Background()
	c, err := compute.NewInstanceGroupsRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	req := &computepb.ListInstancesInstanceGroupsRequest{
		Project:       projectID,
		InstanceGroup: name,
		Zone:          zone,
	}
	it := c.ListInstances(ctx, req)
	var instances []*Instance
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		instances = append(instances, &Instance{
			Name:   path.Base(resp.GetInstance()),
			Zone:   strings.Split(resp.GetInstance(), "/")[8],
			Status: resp.GetStatus(),
		})
	}

	return instances, nil
}

func (b *BackendService) ListServerlesses(projectID string) ([]*Serverless, error) {
	ctx := context.Background()
	c, err := compute.NewRegionNetworkEndpointGroupsRESTClient(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	var serverlesses []*Serverless
	for _, backend := range b.Backends {
		name := path.Base(backend.GetGroup())
		region := strings.Split(backend.GetGroup(), "/")[8]
		groupType := strings.Split(backend.GetGroup(), "/")[9]

		if groupType != "networkEndpointGroups" {
			continue
		}

		req := &computepb.GetRegionNetworkEndpointGroupRequest{
			Project:              projectID,
			NetworkEndpointGroup: name,
			Region:               region,
		}
		resp, err := c.Get(ctx, req)
		if err != nil {
			return nil, err
		}

		if serverless := resp.GetCloudRun(); serverless != nil {
			serverlesses = append(serverlesses, &Serverless{
				Name:    serverless.GetService(),
				Region:  region,
				Service: "Cloud Run",
			})
		}
	}

	return serverlesses, nil
}

func (x *Instance) GetComputeEngine(projectID string) (computeengine.ComputeEngine, error) {
	c, err := computeengine.GetComputeInstance(projectID, x.Zone, x.Name)
	if err != nil {
		return computeengine.ComputeEngine{}, err
	}

	return c, nil
}

func (x *Serverless) Get(projectID string) (computing.Computing, error) {
	switch x.Service {
	case "Cloud Run":
		return cloudrun.GetCloudRunService(projectID, x.Region, x.Name)
	default:
		return nil, fmt.Errorf("%s is not supported service", x.Service)
	}
}

func (r LoadBalancingHTTPS) GetID() string {
	return r.id
}

func (r LoadBalancingHTTPS) GetBackends() []computing.Computing {
	return r.backends
}
