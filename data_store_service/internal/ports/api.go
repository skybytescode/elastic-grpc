package ports

import (
	"context"

	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
)

type APIPort interface {
	// PlaceData copies the ingest worker's ads into Elasticsearch and returns
	// how many were indexed.
	PlaceData(ctx context.Context) (int, error)
	GetAllData(ctx context.Context) ([]domain.Adv, error)
	TextSearch(ctx context.Context, keyword string) ([]domain.Adv, error)
	ScrollSearch(ctx context.Context, from int, size int) ([]domain.Adv, error)
	AggregateSubcategory(ctx context.Context) (map[string]int, error)
}
