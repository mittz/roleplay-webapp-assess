package architecture

import (
	"fmt"
	"log"
	"net/url"

	"github.com/mittz/roleplay-webapp-assess/architecture/computing"
	"github.com/mittz/roleplay-webapp-assess/architecture/computing/appengine"
	"github.com/mittz/roleplay-webapp-assess/architecture/computing/cloudfunctions"
	"github.com/mittz/roleplay-webapp-assess/architecture/computing/cloudrun"
	"github.com/mittz/roleplay-webapp-assess/architecture/computing/computeengine"
	"github.com/mittz/roleplay-webapp-assess/architecture/database"
	"github.com/mittz/roleplay-webapp-assess/architecture/database/alloydb"
	"github.com/mittz/roleplay-webapp-assess/architecture/database/cloudspanner"
	"github.com/mittz/roleplay-webapp-assess/architecture/database/cloudsql"
	"github.com/mittz/roleplay-webapp-assess/architecture/loadbalancing"
	"github.com/mittz/roleplay-webapp-assess/utils"
)

type Architecture struct {
	lb   loadbalancing.LoadBalancingHTTPS
	apps []computing.Computing
	db   database.Database
}

func NewArchitecture(projectID string, endpoint string) Architecture {
	arch := Architecture{}

	u, err := url.Parse(endpoint)
	if err != nil {
		log.Println(err)
		return Architecture{}
	}
	host := u.Host

	if lb, ok := loadbalancing.GetLoadBalancingHTTPS(projectID, host); ok {
		arch.lb = lb
		log.Printf("Load Balancing resource was found: %s", lb.GetID())

		arch.apps = arch.lb.GetBackends()
	} else {
		if computing, ok := computeengine.GetComputeEngine(projectID, host); ok {
			arch.apps = append(arch.apps, computing)
			log.Printf("Compute Engine resource was found: %s", computing.GetID())
		} else if computing, ok := appengine.GetAppEngine(projectID, host); ok {
			arch.apps = append(arch.apps, computing)
			log.Printf("App Engine resource was found: %s", computing.GetID())
		} else if computing, ok := cloudrun.GetCloudRun(projectID, host); ok {
			arch.apps = append(arch.apps, computing)
			log.Printf("Cloud Run resource was found: %s", computing.GetID())
		} else if computing, ok := cloudfunctions.GetCloudFunctions(projectID, host); ok {
			arch.apps = append(arch.apps, computing)
			log.Printf("Cloud Functions resource was found: %s", computing.GetID())
		} else {
			log.Printf("Computing resource (ProjectID: %s, Host: %s) was not found.", projectID, host)
		}
	}

	if db, ok := cloudsql.GetCloudSQL(projectID); ok {
		arch.db = db
		log.Printf("Cloud SQL resource was found: %s", db.GetID())
	} else if db, ok := alloydb.GetAlloyDB(projectID); ok {
		arch.db = db
		log.Printf("AlloyDB resource was found: %s", db.GetID())
	} else if db, ok := cloudspanner.GetCloudSpanner(projectID); ok {
		arch.db = db
		log.Printf("Cloud Spanner resource was found: %s", db.GetID())
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
		case computeengine.ComputeEngine:
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

	log.Printf("app rate: %d, db rate: %d", appRate, dbRate)

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
