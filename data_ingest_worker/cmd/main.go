package main

import (
	"log"

	"github.com/skybytescode/elastic-grpc/data_ingest_worker/adapters/db"
	"github.com/skybytescode/elastic-grpc/data_ingest_worker/adapters/grpc"
	"github.com/skybytescode/elastic-grpc/data_ingest_worker/config"
	"github.com/skybytescode/elastic-grpc/data_ingest_worker/internal/application/core/api"
)

func main() {
	dbAdapter, err := db.NewAdapter(config.GetDataFile())
	if err != nil {
		log.Fatalf("Failed to load the ads: %v", err)
	}
	application := api.NewApplication(dbAdapter)
	grpc.NewAdapter(application, config.GetApplicationPort()).Run()
}
