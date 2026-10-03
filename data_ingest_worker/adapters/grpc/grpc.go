package grpc

import (
	"context"
	"github.com/skybytescode/elastic-grpc/proto/ingestworker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

func (a *Adapter) GetData(ctx context.Context, empty *ingestworker.Empty) (*ingestworker.GetDataResponse, error) {
	log.Println("Data Retrieving...")
	result, err := a.api.GetData()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve ads: %v", err)
	}
	// Convert []domain.Ad to []*ingestworker.Ad
	var ads []*ingestworker.Ad
	for _, ad := range result {
		ads = append(ads, &ingestworker.Ad{
			XId: ad.ID,
			Categories: &ingestworker.Category{
				Subcategory: ad.Categories.Subcategory,
			},
			Title: &ingestworker.Title{
				Ro: ad.Title.Ro,
				Ru: ad.Title.Ru,
			},
			Type:   ad.Type,
			Posted: ad.Posted,
		})
	}
	return &ingestworker.GetDataResponse{Ads: ads}, nil
}
