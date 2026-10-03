package httpServ

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
	"github.com/skybytescode/elastic-grpc/data_store_service/mocks"
)

// A failed transfer must not be reported as a success.
func TestCreateReportsFailures(t *testing.T) {
	mockAPI := &mocks.MockAPIPort{
		PlaceDataFn: func(ctx context.Context) (int, error) { return 0, errors.New("elasticsearch is down") },
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/create", nil)
	http.HandlerFunc(NewAdapter(mockAPI, 8080).createItemHandler).ServeHTTP(rr, req)

	assert.GreaterOrEqual(t, rr.Code, 500)
	assert.NotContains(t, rr.Body.String(), "successfully")
	assert.Contains(t, rr.Body.String(), "elasticsearch is down")
}

func TestRequestValidation(t *testing.T) {
	never := &mocks.MockAPIPort{} // the handlers must reject these before calling the API
	a := NewAdapter(never, 8080).routes()
	for _, tc := range []struct {
		method, url string
		want        int
	}{
		{http.MethodGet, "/create", http.StatusMethodNotAllowed},
		{http.MethodGet, "/searchTitle", http.StatusBadRequest},
		{http.MethodGet, "/searchTitle?title=%20%20", http.StatusBadRequest},
		{http.MethodGet, "/scroll?from=-1", http.StatusBadRequest},
		{http.MethodGet, "/scroll?size=0", http.StatusBadRequest},
		{http.MethodGet, "/scroll?size=1000", http.StatusBadRequest},
		{http.MethodGet, "/scroll?from=abc", http.StatusBadRequest},
	} {
		rr := httptest.NewRecorder()
		a.ServeHTTP(rr, httptest.NewRequest(tc.method, tc.url, nil))
		assert.Equal(t, tc.want, rr.Code, "%s %s", tc.method, tc.url)
	}
}

func TestSearchFailureIsABadGateway(t *testing.T) {
	mockAPI := &mocks.MockAPIPort{
		TextSearchFn: func(ctx context.Context, title string) ([]domain.Adv, error) {
			return nil, errors.New("index adv not found")
		},
	}
	rr := httptest.NewRecorder()
	NewAdapter(mockAPI, 8080).routes().ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/searchTitle?title=casa", nil))
	assert.Equal(t, http.StatusBadGateway, rr.Code)
	assert.JSONEq(t, `{"error": "index adv not found"}`, rr.Body.String())
}
