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
	arch := architecture.NewArchitecture(projectID, endpoint)

	jobHistory := &database.JobHistory{Userkey: userkey, LDAP: user.GetUser(userkey).LDAP, ExecutedAt: time.Now()}

	availabilityRate, err := arch.CalcAvailabilityRate()
	if err != nil {
		jobHistory.Message = fmt.Sprintf("Failed to get availability rate: %v", err.Error())
		if writeErr := jobHistory.WriteDatabase(); writeErr != nil {
			log.Println(writeErr)
		}
		return
	}
	jobHistory.AvailabilityRate = availabilityRate
	log.Printf("Availability rate: %d", availabilityRate)

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
	jobHistory.Message = "Successfully your assessment was completed."

	if writeErr := jobHistory.WriteDatabase(); writeErr != nil {
		log.Println(writeErr)
	}

	log.Println("Successfully completed.")
}
