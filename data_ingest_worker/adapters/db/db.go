package db

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/skybytescode/elastic-grpc/data_ingest_worker/internal/application/domain"
)

// Adapter serves the ads read from a JSON file at start-up.
type Adapter struct {
	ads []domain.Ad
}

// NewAdapter reads the ads from path once, so a missing or broken file stops
// the worker at start-up instead of failing every request.
func NewAdapter(path string) (*Adapter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ads []domain.Ad
	if err := json.Unmarshal(data, &ads); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}
	return &Adapter{ads: ads}, nil
}

// Get returns a copy of the ads, so callers cannot change what later
// requests receive.
func (a *Adapter) Get() ([]domain.Ad, error) {
	return append([]domain.Ad(nil), a.ads...), nil
}
