package mocks

import (
	"context"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
)

type MockAPIPort struct {
	PlaceDataFn            func(ctx context.Context) error
	GetAllDataFn           func(ctx context.Context) ([]domain.Adv, error)
	TextSearchFn           func(ctx context.Context, title string) ([]domain.Adv, error)
	ScrollSearchFn         func(ctx context.Context, from, size int) ([]domain.Adv, error)
	AggregateSubcategoryFn func(ctx context.Context) (map[string]int, error)
}

func (m *MockAPIPort) PlaceData(ctx context.Context) error {
	return m.PlaceDataFn(ctx)
}

func (m *MockAPIPort) GetAllData(ctx context.Context) ([]domain.Adv, error) {
	return m.GetAllDataFn(ctx)
}

func (m *MockAPIPort) TextSearch(ctx context.Context, title string) ([]domain.Adv, error) {
	return m.TextSearchFn(ctx, title)
}

func (m *MockAPIPort) ScrollSearch(ctx context.Context, from, size int) ([]domain.Adv, error) {
	return m.ScrollSearchFn(ctx, from, size)
}

func (m *MockAPIPort) AggregateSubcategory(ctx context.Context) (map[string]int, error) {
	return m.AggregateSubcategoryFn(ctx)
}

func (m *MockAPIPort) SeedData(ctx context.Context) error {
	return m.SeedData(ctx)
}
