package db

import "testing"

func newTestAdapter(t *testing.T) *Adapter {
	t.Helper()
	a, err := NewAdapter("data.json")
	if err != nil {
		t.Fatal(err)
	}
	return a
}
