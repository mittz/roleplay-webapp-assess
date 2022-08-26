package utils

import (
	"log"
	"os"
)

func getEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Fatalf("%s is not set", key)
	}

	return value
}

func GetEnvUserkey() string {
	return getEnv("USER_KEY")
}

func GetEnvEndpoint() string {
	return getEnv("ENDPOINT")
}

func GetEnvProjectID() string {
	return getEnv("PROJECT_ID")
}

func GetMin(x, y int) int {
	if x < y {
		return x
	}

	return y
}
