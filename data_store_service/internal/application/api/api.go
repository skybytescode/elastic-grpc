package api

import (
	"context"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/ports"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Application struct {
	db         ports.DBPort
	dataworker ports.DataIngestWorkerPort
}

func NewApplication(db ports.DBPort, data_worker ports.DataIngestWorkerPort) *Application {
	return &Application{
		db:         db,
		dataworker: data_worker,
	}
}

func (a Application) PlaceData(ctx context.Context) error {
	dataAds, dateRetrieveErr := a.dataworker.GetDataFromWorker(ctx)
	if dateRetrieveErr != nil {
		st, _ := status.FromError(dateRetrieveErr)
		fieldErr := &errdetails.BadRequest_FieldViolation{
			Field:       "data",
			Description: st.Message(),
		}
		badReq := &errdetails.BadRequest{}
		badReq.FieldViolations = append(badReq.FieldViolations, fieldErr)
		orderStatus := status.New(codes.InvalidArgument, "order creation failed")
		statusWithDetails, _ := orderStatus.WithDetails(badReq)
		return statusWithDetails.Err()
	}
	for _, dataAd := range dataAds {
		a.db.InsertData(ctx, dataAd)
	}
	return nil
}

func (a Application) GetAllData(ctx context.Context) ([]domain.Adv, error) {
	ads, err := a.db.GetAllDocuments()
	if err != nil {
		st, _ := status.FromError(err)
		fieldErr := &errdetails.BadRequest_FieldViolation{
			Field:       "data",
			Description: st.Message(),
		}
		badReq := &errdetails.BadRequest{}
		badReq.FieldViolations = append(badReq.FieldViolations, fieldErr)
		orderStatus := status.New(codes.InvalidArgument, "get data failed")
		statusWithDetails, _ := orderStatus.WithDetails(badReq)
		return nil, statusWithDetails.Err()
	}
	return ads, err
}
func (a Application) TextSearch(ctx context.Context, keyword string) ([]domain.Adv, error) {
	ads, err := a.db.FullTextSearch(keyword)
	if err != nil {
		st, _ := status.FromError(err)
		fieldErr := &errdetails.BadRequest_FieldViolation{
			Field:       "data",
			Description: st.Message(),
		}
		badReq := &errdetails.BadRequest{}
		badReq.FieldViolations = append(badReq.FieldViolations, fieldErr)
		orderStatus := status.New(codes.InvalidArgument, "get data failed")
		statusWithDetails, _ := orderStatus.WithDetails(badReq)
		return nil, statusWithDetails.Err()
	}
	return ads, err
}

func (a Application) ScrollSearch(ctx context.Context, from int, size int) ([]domain.Adv, error) {
	ads, err := a.db.InfiniteScroll(from, size)
	if err != nil {
		st, _ := status.FromError(err)
		fieldErr := &errdetails.BadRequest_FieldViolation{
			Field:       "data",
			Description: st.Message(),
		}
		badReq := &errdetails.BadRequest{}
		badReq.FieldViolations = append(badReq.FieldViolations, fieldErr)
		orderStatus := status.New(codes.InvalidArgument, "get data failed")
		statusWithDetails, _ := orderStatus.WithDetails(badReq)
		return nil, statusWithDetails.Err()
	}
	return ads, err
}

func (a Application) AggregateSubcategory(ctx context.Context) (map[string]int, error) {
	ads, err := a.db.AggregateBySubcategory()
	if err != nil {
		st, _ := status.FromError(err)
		fieldErr := &errdetails.BadRequest_FieldViolation{
			Field:       "data",
			Description: st.Message(),
		}
		badReq := &errdetails.BadRequest{}
		badReq.FieldViolations = append(badReq.FieldViolations, fieldErr)
		orderStatus := status.New(codes.InvalidArgument, "get data failed")
		statusWithDetails, _ := orderStatus.WithDetails(badReq)
		return nil, statusWithDetails.Err()
	}
	return ads, err
}

func (a Application) SeedData(ctx context.Context) error {
	err := a.db.SeedingData(ctx)
	if err != nil {
		st, _ := status.FromError(err)
		fieldErr := &errdetails.BadRequest_FieldViolation{
			Field:       "data",
			Description: st.Message(),
		}
		badReq := &errdetails.BadRequest{}
		badReq.FieldViolations = append(badReq.FieldViolations, fieldErr)
		orderStatus := status.New(codes.InvalidArgument, "get data failed")
		statusWithDetails, _ := orderStatus.WithDetails(badReq)
		return statusWithDetails.Err()
	}
	return err
}
