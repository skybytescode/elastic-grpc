package ports

import (
	"context"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
)

type DataIngestWorkerPort interface {
	GetDataFromWorker(context.Context) ([]domain.Adv, error)
}
