package db

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// Each request must return the file's ads once; the adapter used to append
// the file to what it returned before, so every call grew the list.
func TestGetReturnsTheSameAdsEveryTime(t *testing.T) {
	a := newTestAdapter(t)
	first, err := a.Get()
	require.NoError(t, err)
	require.Equal(t, 50, len(first))

	second, err := a.Get()
	require.NoError(t, err)
	require.Equal(t, 50, len(second))
}

func TestMissingOrBrokenFileFailsAtStartUp(t *testing.T) {
	_, err := NewAdapter("does-not-exist.json")
	require.Error(t, err)

	broken := t.TempDir() + "/broken.json"
	require.NoError(t, os.WriteFile(broken, []byte(`[{"_id": `), 0o600))
	_, err = NewAdapter(broken)
	require.ErrorContains(t, err, "decoding")
}
