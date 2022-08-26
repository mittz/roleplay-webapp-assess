package main

import (
	"log"

	"github.com/mittz/roleplay-webapp-assess/architecture"
	"github.com/mittz/roleplay-webapp-assess/benchmark"
	"github.com/mittz/roleplay-webapp-assess/database"
	"github.com/mittz/roleplay-webapp-assess/utils"
)

func main() {
	userkey := utils.GetEnvUserkey()
	endpoint := utils.GetEnvEndpoint()
	projectID := utils.GetEnvProjectID()
	arch := architecture.NewArchitecture(projectID, endpoint)

	jobHistory := new(database.JobHistory)

	availabilityRate, err := arch.CalcAvailabilityRate()
	if err != nil {
		jobHistory.Message = err.Error()
		if writeErr := jobHistory.WriteDatabase(); writeErr != nil {
			log.Println(writeErr)
		}
		return
	}
	jobHistory.AvailabilityRate = availabilityRate

	jobHistory.Cost = arch.CalcCost()

	performance, err := benchmark.Run(userkey, endpoint)
	if err != nil {
		jobHistory.Message = err.Error()
		if writeErr := jobHistory.WriteDatabase(); writeErr != nil {
			log.Println(writeErr)
		}
		return
	}
	jobHistory.Performance = performance

	jobHistory.Total = jobHistory.Performance * jobHistory.AvailabilityRate
	jobHistory.CostPerformance = float64(jobHistory.Total) / jobHistory.Cost
	jobHistory.Message = "Successfully your assessment was completed."

	if writeErr := jobHistory.WriteDatabase(); writeErr != nil {
		log.Println(writeErr)
	}
}
