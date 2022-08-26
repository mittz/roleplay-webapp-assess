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
	host := u.Host

	if lb, ok := GetLoadBalancingHTTPS(projectID, host); ok {
		arch.lb = lb
		arch.apps = arch.lb.GetBackends()
	} else {
		log.Printf("Load Balancing (ProjectID: %s, Host: %s) resource is not found.", projectID, host)

		if computing, ok := GetComputeEngine(projectID, host); ok {
			arch.apps = append(arch.apps, computing)
		} else if computing, ok := GetAppEngine(projectID, host); ok {
			arch.apps = append(arch.apps, computing)
		} else if computing, ok := GetCloudRun(projectID, host); ok {
			arch.apps = append(arch.apps, computing)
		} else if computing, ok := GetCloudFunctions(projectID, host); ok {
			arch.apps = append(arch.apps, computing)
		} else if computing, ok := GetKubernetesEngine(projectID, host); ok {
			arch.apps = append(arch.apps, computing)
		} else {
			log.Printf("Computing (ProjectID: %s, Host: %s) resource is not found.", projectID, host)
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

	includeServerless := false
	for _, app := range a.apps {
		appRegions[app.GetRegion()] = struct{}{}
		switch app.(type) {
		case ComputeEngine:
			appZones[app.GetZone()] = struct{}{}
		default:
			// For serverless services
			includeServerless = true
		}
	}

	if len(appRegions) > 1 {
		appRate = 3
	} else if includeServerless {
		appRate = 2
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
