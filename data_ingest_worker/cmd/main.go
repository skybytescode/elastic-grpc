package main

import (
	"github.com/skybytescode/elastic-grpc/data_ingest_worker/adapters/db"
	"github.com/skybytescode/elastic-grpc/data_ingest_worker/adapters/grpc"
	"github.com/skybytescode/elastic-grpc/data_ingest_worker/config"
	"github.com/skybytescode/elastic-grpc/data_ingest_worker/internal/application/core/api"
	"log"
	"os"
)

func main() {

	err := os.Setenv("APPLICATION_PORT", "4001")
	if err != nil {
		return
	}
	err = os.Setenv("ENV", "development")
	if err != nil {
		return
	}

	dbAdapter, err := db.NewAdapter()
	if err != nil {
		log.Fatalf("Failed to connect to database. Error: %v", err)
	}

	application := api.NewApplication(dbAdapter)
	grpcAdapter := grpc.NewAdapter(application, config.GetApplicationPort())
	grpcAdapter.Run()
}
