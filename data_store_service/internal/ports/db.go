package ports

import (
	"context"

	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
)

type DBPort interface {
	InsertAll(ctx context.Context, ads []domain.Adv) (int, error)
	GetAllDocuments(ctx context.Context) ([]domain.Adv, error)
	FullTextSearch(ctx context.Context, keyword string) ([]domain.Adv, error)
	InfiniteScroll(ctx context.Context, from, size int) ([]domain.Adv, error)
	AggregateBySubcategory(ctx context.Context) (map[string]int, error)
}
