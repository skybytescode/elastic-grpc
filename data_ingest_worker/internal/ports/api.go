package ports

import (
	"github.com/skybytescode/elastic-grpc/data_ingest_worker/internal/application/domain"
)

type APIPort interface {
	GetData() ([]domain.Ad, error)
}
