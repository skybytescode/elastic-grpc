package main

import (
	"github.com/skybytescode/elastic-grpc/data_store_service/config"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/adapters/dataworker"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/adapters/db"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/adapters/httpServ"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/api"
	"log"
	"os"
)

func main() {
	err := os.Setenv("APPLICATION_PORT", "8080")
	if err != nil {
		return
	}
	err = os.Setenv("ENV", "development")
	if err != nil {
		return
	}
	//err = os.Setenv("DATA_INGEST_WORKER_URL", "127.0.0.1:4001")
	err = os.Setenv("DATA_INGEST_WORKER_URL", "grpcserver:4001")
	if err != nil {
		return
	}

	dbAdapter, err := db.NewAdapter()
	if err != nil {
		log.Fatalf("Failed to connect to database. Error: %v", err)
	}
	dataWorkerAdaptor, err := dataworker.NewAdapter(config.GetDataIngestWorkerUrl())
	if err != nil {
		log.Fatalf("Failed to initialize payment stub. Error: %v", err)
	}
	application := api.NewApplication(dbAdapter, dataWorkerAdaptor)
	httpAdapter := httpServ.NewAdapter(application, config.GetApplicationPort())
	httpAdapter.Run()
}
