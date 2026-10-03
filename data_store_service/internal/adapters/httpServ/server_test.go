package httpServ

import (
	"context"
	"encoding/json"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
	"github.com/skybytescode/elastic-grpc/data_store_service/mocks"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateItemHandler(t *testing.T) {
	mockAPI := &mocks.MockAPIPort{
		PlaceDataFn: func(ctx context.Context) (int, error) {
			return 50, nil
		},
	}
	adapter := NewAdapter(mockAPI, 8080)

	req, err := http.NewRequest("POST", "/create", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.createItemHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"indexed": 50}`, strings.TrimSpace(rr.Body.String()))
}

func TestGetAllDocsHandler(t *testing.T) {
	mockAPI := &mocks.MockAPIPort{
		GetAllDataFn: func(ctx context.Context) ([]domain.Adv, error) {
			return []domain.Adv{
				{
					ID: "1",
					Categories: domain.Category{
						Subcategory: "1401",
					},
					Title: domain.Title{
						Ro: "title_Ro",
						Ru: "title_Ru",
					},
					Type:   "standard",
					Posted: 1486556302.101039,
				},
				{
					ID: "2",
					Categories: domain.Category{
						Subcategory: "1401",
					},
					Title: domain.Title{
						Ro: "title_Ro",
						Ru: "title_Ru",
					},
					Type:   "standard",
					Posted: 1486556302.101039,
				},
			}, nil
		},
	}
	adapter := NewAdapter(mockAPI, 8080)

	req, err := http.NewRequest("GET", "/getalldocs", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.getAllDocsHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []domain.Adv
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
}

func TestSearchByTitleHandler(t *testing.T) {
	mockAPI := &mocks.MockAPIPort{
		TextSearchFn: func(ctx context.Context, title string) ([]domain.Adv, error) {
			return []domain.Adv{
				{
					ID: "1",
					Categories: domain.Category{
						Subcategory: "1401",
					},
					Title: domain.Title{
						Ro: "title_Ro",
						Ru: "title_Ru",
					},
					Type:   "standard",
					Posted: 1486556302.101039,
				},
			}, nil
		},
	}
	adapter := NewAdapter(mockAPI, 8080)

	req, err := http.NewRequest("GET", "/searchTitle?title=Doc1", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.searchByTitle)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []domain.Adv
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
}

func TestScrollSearchHandler(t *testing.T) {
	mockAPI := &mocks.MockAPIPort{
		ScrollSearchFn: func(ctx context.Context, from, size int) ([]domain.Adv, error) {
			return []domain.Adv{
				{
					ID: "1",
					Categories: domain.Category{
						Subcategory: "1401",
					},
					Title: domain.Title{
						Ro: "title_Ro",
						Ru: "title_Ru",
					},
					Type:   "standard",
					Posted: 1486556302.101039,
				},
				{
					ID: "2",
					Categories: domain.Category{
						Subcategory: "1401",
					},
					Title: domain.Title{
						Ro: "title_Ro",
						Ru: "title_Ru",
					},
					Type:   "standard",
					Posted: 1486556302.101039,
				},
			}, nil
		},
	}
	adapter := NewAdapter(mockAPI, 8080)

	req, err := http.NewRequest("GET", "/scroll?from=0&size=2", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.scrollSearch)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []domain.Adv
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
}

func TestAggSubcategoryHandler(t *testing.T) {
	mockAPI := &mocks.MockAPIPort{
		AggregateSubcategoryFn: func(ctx context.Context) (map[string]int, error) {
			return map[string]int{
				"sub1": 2,
				"sub2": 3,
			}, nil
		},
	}
	adapter := NewAdapter(mockAPI, 8080)

	req, err := http.NewRequest("GET", "/aggsub", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(adapter.aggSubcategory)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]int
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 2, response["sub1"])
	assert.Equal(t, 3, response["sub2"])
}
