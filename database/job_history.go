package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4"
)

type JobHistory struct {
	ID               int
	Userkey          string
	LDAP             string
	Score            int
	ScoreByCost      float64
	Performance      int
	AvailabilityRate int
	Message          string
	Cost             float64
	ExecutedAt       time.Time
}

func (j JobHistory) WriteDatabase() error {
	dp := GetDatabaseConnection()
	queryInsertHistory := `
		INSERT INTO job_histories(
			userkey,
			ldap,
			score,
			score_by_cost,
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
	if _, err := dp.Exec(context.Background(), queryInsertHistory,
		j.Userkey,
		j.LDAP,
		j.Score,
		j.ScoreByCost,
		j.Performance,
		j.AvailabilityRate,
		j.Message,
		j.Cost,
		j.ExecutedAt,
	); err != nil {
		return err
	}

	var ldap string
	queryGetRanking := `
		SELECT ldap FROM rankings WHERE ldap=$1
	`
	queryInsertRanking := `
		INSERT INTO rankings(
			ldap,
			score,
			score_by_cost,
			executed_at
		) VALUES(
			$1,
			$2,
			$3,
			$4
		)
	`
	queryUpdateRanking := `
		UPDATE rankings SET total=$1, cost_performance=$2, executed_at=$3 WHERE ldap=$4
	`

	isRowPresent := true
	if err := dp.QueryRow(context.Background(), queryGetRanking, j.LDAP).Scan(&ldap); err != nil {
		if err != pgx.ErrNoRows {
			return err
		}

		isRowPresent = false
	}

	if isRowPresent {
		if _, err := dbPool.Exec(context.Background(), queryUpdateRanking, j.Score, j.ScoreByCost, j.ExecutedAt, j.LDAP); err != nil {
			return err
		}
	} else {
		if _, err := dbPool.Exec(context.Background(), queryInsertRanking, j.LDAP, j.Score, j.ScoreByCost, j.ExecutedAt); err != nil {
			return err
		}
	}

	return nil
}
