package dataworker

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
	"github.com/skybytescode/elastic-grpc/proto/ingestworker"
)

type Adapter struct {
	dataworker ingestworker.RetrieveDataClient
}

func NewAdapter(dataWorkerServiceUrl string) (*Adapter, error) {
	conn, err := grpc.NewClient(dataWorkerServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Adapter{dataworker: ingestworker.NewRetrieveDataClient(conn)}, nil
}

func (a *Adapter) GetDataFromWorker(ctx context.Context) ([]domain.Adv, error) {
	data, err := a.dataworker.GetData(ctx, &ingestworker.Empty{})
	if err != nil {
		return nil, err
	}
	ads := make([]domain.Adv, 0, len(data.Ads))
	for _, ad := range data.Ads {
		ads = append(ads, domain.Adv{
			ID:         ad.XId,
			Categories: domain.Category{Subcategory: ad.GetCategories().GetSubcategory()},
			Title:      domain.Title{Ro: ad.GetTitle().GetRo(), Ru: ad.GetTitle().GetRu()},
			Type:       ad.Type,
			Posted:     ad.Posted,
		})
	}
	return ads, nil
}
