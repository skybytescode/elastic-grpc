package main

import (
	"context"
	"log"
	"time"

	"github.com/skybytescode/elastic-grpc/data_store_service/config"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/adapters/dataworker"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/adapters/db"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/adapters/httpServ"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/api"
)

func main() {
	dbAdapter, err := db.NewAdapter(context.Background(), db.Config{
		URL:      config.GetElasticsearchURL(),
		Username: config.GetElasticsearchUsername(),
		Password: config.GetElasticsearchPassword(),
	}, 2*time.Minute) // Elasticsearch can take a while to start
	if err != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}
	dataWorkerAdapter, err := dataworker.NewAdapter(config.GetDataIngestWorkerUrl())
	if err != nil {
		log.Fatalf("Failed to create the ingest worker client: %v", err)
	}
	application := api.NewApplication(dbAdapter, dataWorkerAdapter)
	httpServ.NewAdapter(application, config.GetApplicationPort()).Run()
}
