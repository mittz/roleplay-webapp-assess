package main

import (
	"fmt"
	"log"
	"time"

	"github.com/mittz/roleplay-webapp-assess/architecture"
	"github.com/mittz/roleplay-webapp-assess/benchmark"
	"github.com/mittz/roleplay-webapp-assess/database"
	"github.com/mittz/roleplay-webapp-assess/user"
	"github.com/mittz/roleplay-webapp-assess/utils"
)

func main() {
	userkey := utils.GetEnvUserkey()
	endpoint := utils.GetEnvEndpoint()
	projectID := utils.GetEnvProjectID()
	jobHistory := &database.JobHistory{Userkey: userkey, LDAP: user.GetUser(userkey).LDAP, ExecutedAt: time.Now()}

	arch, err := architecture.NewArchitecture(projectID, endpoint)
	if err != nil {
		jobHistory.Message = fmt.Sprintf("Failed to get architecture information: %v", err.Error())
		if writeErr := jobHistory.WriteDatabase(); writeErr != nil {
			log.Println(writeErr)
		}
		return
	}

	availabilityRates, err := arch.CalcAvailabilityRate()
	if availabilityRates == nil || len(availabilityRates) < 2 || err != nil {
		jobHistory.AvailabilityRate = 0
		jobHistory.Message = fmt.Sprintf("Failed to get availability rate: %v", err.Error())
		if writeErr := jobHistory.WriteDatabase(); writeErr != nil {
			log.Println(writeErr)
		}
		return
	}
	appRate, dbRate := availabilityRates[0], availabilityRates[1]
	jobHistory.AvailabilityRate = utils.GetMin(appRate, dbRate)

	jobHistory.Cost = arch.CalcCost()

	performance, err := benchmark.Run(userkey, endpoint)
	if err != nil {
		jobHistory.Message = fmt.Sprintf("Failed to get benchmark score: %v", err.Error())
		if writeErr := jobHistory.WriteDatabase(); writeErr != nil {
			log.Println(writeErr)
		}
		return
	}
	jobHistory.Performance = performance

	jobHistory.Score = jobHistory.Performance * jobHistory.AvailabilityRate
	jobHistory.ScoreByCost = float64(jobHistory.Score) / jobHistory.Cost
	jobHistory.Message = fmt.Sprintf("Successfully your assessment was completed. App rate: %d DB rate: %d", appRate, dbRate)

	if writeErr := jobHistory.WriteDatabase(); writeErr != nil {
		log.Println(writeErr)
	}

	log.Printf("Successfully your assessment was completed. App rate: %d DB rate: %d", appRate, dbRate)
}
