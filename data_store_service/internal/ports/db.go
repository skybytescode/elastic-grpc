package ports

import (
	"context"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
)

type DBPort interface {
	SeedingData(ctx context.Context) error
	InsertData(ctx context.Context, adv domain.Adv) error
	GetAllDocuments() ([]domain.Adv, error)
	FullTextSearch(keyword string) ([]domain.Adv, error)
	InfiniteScroll(from int, size int) ([]domain.Adv, error)
	AggregateBySubcategory() (map[string]int, error)
}
