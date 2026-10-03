// Package config reads the worker settings from the environment, with
// defaults that work when running from the data_ingest_worker folder.
package config

import (
	"log"
	"os"
	"strconv"
)

func GetEnv() string { return envOr("ENV", "development") }

func GetApplicationPort() int {
	port, err := strconv.Atoi(envOr("APPLICATION_PORT", "4001"))
	if err != nil {
		log.Fatalf("APPLICATION_PORT is not a number: %v", err)
	}
	return port
}

// GetDataFile is the JSON file with the ads to serve.
func GetDataFile() string { return envOr("DATA_FILE", "adapters/db/data.json") }

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
