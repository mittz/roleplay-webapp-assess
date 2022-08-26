package architecture

import (
	"fmt"
	"log"
	"net/url"

	"github.com/mittz/roleplay-webapp-assess/utils"
)

type Architecture struct {
	lb   LoadBalancingHTTPS
	apps []Computing
	db   Database
}

func NewArchitecture(projectID string, endpoint string) Architecture {
	arch := Architecture{}

	u, err := url.Parse(endpoint)
	if err != nil {
	}
	hostIP := u.Host

	if lb, ok := GetLoadBalancingHTTPS(projectID, hostIP); ok {
		arch.lb = lb
		arch.apps = arch.lb.GetBackends()
	} else {
		log.Printf("Load Balancing (ProjectID: %s, HostIP: %s) resource is not found.", projectID, hostIP)

		if computing, ok := GetComputeEngine(projectID, hostIP); ok {
			arch.apps = append(arch.apps, computing)
		} else if computing, ok := GetAppEngine(projectID, hostIP); ok {
			arch.apps = append(arch.apps, computing)
		} else if computing, ok := GetCloudRun(projectID, hostIP); ok {
			arch.apps = append(arch.apps, computing)
		} else if computing, ok := GetCloudFunctions(projectID, hostIP); ok {
			arch.apps = append(arch.apps, computing)
		} else if computing, ok := GetKubernetesEngine(projectID, hostIP); ok {
			arch.apps = append(arch.apps, computing)
		} else {
			log.Printf("Computing (ProjectID: %s, HostIP: %s) resource is not found.", projectID, hostIP)
		}
	}

	if db, ok := GetCloudSQL(projectID); ok {
		arch.db = db
	} else if db, ok := GetAlloyDB(projectID); ok {
		arch.db = db
	} else if db, ok := GetCloudSpanner(projectID); ok {
		arch.db = db
	} else {
		log.Printf("Database (ProjectID: %s) resource is not found.", projectID)
	}

	return arch
}

func (a Architecture) CalcAvailabilityRate() (int, error) {
	if len(a.apps) == 0 {
		return 0, fmt.Errorf("Computing product was not found")
	}

	var appRate, dbRate int
	appRegions, appZones := make(map[string]interface{}), make(map[string]interface{})

	for _, app := range a.apps {
		if region := app.GetRegion(); region != "" {
			appRegions[region] = struct{}{}
		}

		if zone := app.GetZone(); zone != "" {
			appZones[zone] = struct{}{}
		}
	}

	if len(appRegions) > 1 {
		appRate = 3
	} else if len(appZones) > 1 {
		appRate = 2
	} else if len(appZones) == 1 {
		appRate = 1
	} else {
		appRate = 0
	}

	if a.db == nil {
		return 0, fmt.Errorf("Database product was not found")
	}
	dbRate = a.db.GetAvailabilityRate()

	return utils.GetMin(appRate, dbRate), nil
}

func (a Architecture) CalcCost() float64 {
	total := 0.0

	for _, app := range a.apps {
		total += app.GetCost()
	}

	if a.db != nil {
		total += a.db.GetCost()
	}

	return total
}
