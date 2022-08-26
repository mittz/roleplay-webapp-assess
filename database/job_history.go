package database

import (
	"context"
	"time"
)

type JobHistory struct {
	ID               int
	Userkey          string
	LDAP             string
	Total            int
	CostPerformance  float64
	Performance      int
	AvailabilityRate int
	Message          string
	Cost             float64
	ExecutedAt       time.Time
}

func (j JobHistory) WriteDatabase() error {
	dp := GetDatabaseConnection()
	query := `
		INSERT INTO job_histories(
			userkey,
			ldap,
			total,
			cost_performance,
			performance,
			availability_rate,
			message,
			cost,
			executed_at
		) VALUES(
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9
		)
	`
	_, err := dp.Exec(context.Background(), query,
		j.Userkey,
		j.LDAP,
		j.Total,
		j.CostPerformance,
		j.Performance,
		j.AvailabilityRate,
		j.Message,
		j.Cost,
		time.Now(),
	)

	return err
}
