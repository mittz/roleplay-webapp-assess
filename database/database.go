package database

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"cloud.google.com/go/cloudsqlconn"
	"github.com/jackc/pgx/v4/pgxpool"
)

var dbPool *pgxpool.Pool

func getInstanceConnectionName() string {
	value, ok := os.LookupEnv("INSTANCE_CONNECTION_NAME")
	if !ok {
		log.Fatal("Env: INSTANCE_CONNECTION_NAME must be set.")
	}

	return value
}

func getDatabaseName() string {
	value, ok := os.LookupEnv("DATABASE_NAME")
	if !ok {
		log.Fatal("Env: DATABASE_NAME must be set.")
	}

	return value
}

func getDatabaseUser() string {
	value, ok := os.LookupEnv("DATABASE_USER")
	if !ok {
		log.Fatal("Env: DATABASE_USER must be set.")
	}

	return value
}

func GetDatabaseConnection() *pgxpool.Pool {
	if dbPool != nil {
		return dbPool
	}

	ctx := context.Background()

	dsn := fmt.Sprintf("user=%s password=\"\" dbname=%s sslmode=disable", getDatabaseUser(), getDatabaseName())
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("failed to parse pgx config: %v", err)
	}

	d, err := cloudsqlconn.NewDialer(ctx, cloudsqlconn.WithIAMAuthN())
	if err != nil {
		log.Fatalf("failed to initiate Dialer: %v", err)
	}

	config.ConnConfig.DialFunc = func(ctx context.Context, network string, instance string) (net.Conn, error) {
		return d.Dial(ctx, getInstanceConnectionName())
	}

	pool, connErr := pgxpool.ConnectConfig(ctx, config)
	if connErr != nil {
		log.Fatalf("failed to connect: %s", connErr)
	}
	dbPool = pool

	return dbPool
}
