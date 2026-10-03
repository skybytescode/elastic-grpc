package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/skybytescode/elastic-grpc/data_ingest_worker/adapters/db"
	"github.com/skybytescode/elastic-grpc/data_ingest_worker/internal/application/core/api"
	"github.com/skybytescode/elastic-grpc/proto/ingestworker"
)

// The worker served over real gRPC, called the way the store service calls it.
func TestGetDataOverGRPC(t *testing.T) {
	store, err := db.NewAdapter("../db/data.json")
	require.NoError(t, err)
	worker := NewAdapter(api.NewApplication(store), 0)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go worker.Serve(lis)
	t.Cleanup(worker.Stop)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()
	client := ingestworker.NewRetrieveDataClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for i := 0; i < 2; i++ { // the second call must not return duplicates
		res, err := client.GetData(ctx, &ingestworker.Empty{})
		require.NoError(t, err)
		require.Equal(t, 50, len(res.Ads))
		first := res.Ads[0]
		require.Equal(t, "38118545", first.XId)
		require.Equal(t, "1407", first.GetCategories().GetSubcategory())
		require.Equal(t, "Teren sub constructie in apropiere de Vadul lui Voda", first.GetTitle().GetRo())
	}
}
