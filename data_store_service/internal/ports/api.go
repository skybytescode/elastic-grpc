package ports

import (
	"context"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
)

type APIPort interface {
	SeedData(ctx context.Context) error
	PlaceData(ctx context.Context) error
	GetAllData(ctx context.Context) ([]domain.Adv, error)
	TextSearch(ctx context.Context, keyword string) ([]domain.Adv, error)
	ScrollSearch(ctx context.Context, from int, size int) ([]domain.Adv, error)
	AggregateSubcategory(ctx context.Context) (map[string]int, error)
}
