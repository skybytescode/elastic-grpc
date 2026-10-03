package dataworker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// When the ingest worker is unreachable the store service must get an error,
// not a nil-pointer panic.
func TestUnreachableWorkerIsAnError(t *testing.T) {
	a, err := NewAdapter("127.0.0.1:1") // nothing listens on port 1
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = a.GetDataFromWorker(ctx)
	require.Error(t, err)
}
