// Package config reads the service settings from the environment. Every
// setting has a default that works for running everything on localhost;
// docker-compose.yml sets the container addresses.
package config

import (
	"log"
	"os"
	"strconv"
)

func GetApplicationPort() int {
	port, err := strconv.Atoi(envOr("APPLICATION_PORT", "8080"))
	if err != nil {
		log.Fatalf("APPLICATION_PORT is not a number: %v", err)
	}
	return port
}

func GetDataIngestWorkerUrl() string   { return envOr("DATA_INGEST_WORKER_URL", "localhost:4001") }
func GetElasticsearchURL() string      { return envOr("ELASTICSEARCH_URL", "http://localhost:9200") }
func GetElasticsearchUsername() string { return os.Getenv("ELASTICSEARCH_USERNAME") }
func GetElasticsearchPassword() string { return os.Getenv("ELASTICSEARCH_PASSWORD") }

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
