package api

import (
	"context"
	"fmt"

	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/ports"
)

type Application struct {
	db         ports.DBPort
	dataworker ports.DataIngestWorkerPort
}

func NewApplication(db ports.DBPort, dataWorker ports.DataIngestWorkerPort) *Application {
	return &Application{db: db, dataworker: dataWorker}
}

func (a Application) PlaceData(ctx context.Context) (int, error) {
	ads, err := a.dataworker.GetDataFromWorker(ctx)
	if err != nil {
		return 0, fmt.Errorf("getting ads from the ingest worker: %w", err)
	}
	n, err := a.db.InsertAll(ctx, ads)
	if err != nil {
		return 0, fmt.Errorf("storing ads: %w", err)
	}
	return n, nil
}

func (a Application) GetAllData(ctx context.Context) ([]domain.Adv, error) {
	return a.db.GetAllDocuments(ctx)
}

func (a Application) TextSearch(ctx context.Context, keyword string) ([]domain.Adv, error) {
	return a.db.FullTextSearch(ctx, keyword)
}

func (a Application) ScrollSearch(ctx context.Context, from int, size int) ([]domain.Adv, error) {
	return a.db.InfiniteScroll(ctx, from, size)
}

func (a Application) AggregateSubcategory(ctx context.Context) (map[string]int, error) {
	return a.db.AggregateBySubcategory(ctx)
}
