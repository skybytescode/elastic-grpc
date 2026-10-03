package dataworker

import (
	"context"
	"github.com/skybytescode/elastic-grpc/proto/ingestworker"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

type Adapter struct {
	dataworker ingestworker.RetrieveDataClient
}

func NewAdapter(dataWorkerServiceUrl string) (*Adapter, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.NewClient(dataWorkerServiceUrl, opts...)
	if err != nil {
		return nil, err
	}
	client := ingestworker.NewRetrieveDataClient(conn)
	return &Adapter{dataworker: client}, nil
}

func (a *Adapter) GetDataFromWorker(ctx context.Context) ([]domain.Adv, error) {
	data, err := a.dataworker.GetData(ctx, &ingestworker.Empty{})
	if err != nil {
		log.Printf("error getting data from dataworker: %v", err)
	}
	var domainAds []domain.Adv
	for _, ad := range data.Ads {
		domainAd := domain.Adv{
			ID: ad.XId,
			Categories: domain.Category{
				Subcategory: ad.Categories.Subcategory,
			},
			Title: domain.Title{
				Ro: ad.Title.Ro,
				Ru: ad.Title.Ru,
			},
			Type:   ad.Type,
			Posted: ad.Posted,
		}
		domainAds = append(domainAds, domainAd)
	}
	return domainAds, nil
}
