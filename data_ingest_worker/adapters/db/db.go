package db

import (
	"encoding/json"
	"fmt"
	"github.com/skybytescode/elastic-grpc/data_ingest_worker/internal/application/domain"
	"os"
)

type Adapter struct {
	db *[]domain.Ad
}

func NewAdapter() (*Adapter, error) {
	var advData []domain.Ad
	adapter := &Adapter{
		db: &advData,
	}
	return adapter, nil
}

func (a *Adapter) Get() ([]domain.Ad, error) {
	//	file, err := os.Open("adapters/db/data.json") // using localhost with make run
	file, err := os.Open("data.json") // using docker container
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, err
	}
	defer file.Close()

	// Decode the JSON file into a slice of domain.Ad
	var ads []domain.Ad
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&ads); err != nil {
		fmt.Println("Error decoding JSON:", err)
		return nil, err
	}
	*a.db = append(*a.db, ads...)
	return *a.db, nil
}
