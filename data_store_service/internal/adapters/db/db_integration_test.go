package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
)

// These tests need a running Elasticsearch, for example
//
//	docker run -p 9200:9200 -e discovery.type=single-node -e xpack.security.enabled=false elasticsearch:8.19.22
//
// and ELASTICSEARCH_TEST_URL=http://localhost:9200. Without it they are skipped.
// Each test gets its own index, deleted afterwards.
func newTestAdapter(t *testing.T) *Adapter {
	t.Helper()
	url := os.Getenv("ELASTICSEARCH_TEST_URL")
	if url == "" {
		t.Skip("ELASTICSEARCH_TEST_URL is not set; skipping Elasticsearch integration test")
	}
	ctx := context.Background()
	index := fmt.Sprintf("adv_test_%d", time.Now().UnixNano())
	a, err := NewAdapter(ctx, Config{URL: url, Index: index}, 30*time.Second)
	require.NoError(t, err)
	t.Cleanup(func() { _ = a.do(ctx, http.MethodDelete, "/"+index, nil, nil) })
	return a
}

// sampleAds is the ingest worker's real data: 50 Moldovan classified ads with
// Romanian and Russian titles.
func sampleAds(t *testing.T) []domain.Adv {
	t.Helper()
	data, err := os.ReadFile("../../../../data_ingest_worker/adapters/db/data.json")
	require.NoError(t, err)
	var raw []struct {
		ID         string          `json:"_id"`
		Categories domain.Category `json:"categories"`
		Title      domain.Title    `json:"title"`
		Type       string          `json:"type"`
		Posted     float64         `json:"posted"`
	}
	require.NoError(t, json.Unmarshal(data, &raw))
	ads := make([]domain.Adv, len(raw))
	for i, r := range raw {
		ads[i] = domain.Adv{ID: r.ID, Categories: r.Categories, Title: r.Title, Type: r.Type, Posted: r.Posted}
	}
	return ads
}

func loaded(t *testing.T) (*Adapter, []domain.Adv) {
	t.Helper()
	a := newTestAdapter(t)
	ads := sampleAds(t)
	n, err := a.InsertAll(context.Background(), ads)
	require.NoError(t, err)
	require.Equal(t, len(ads), n)
	return a, ads
}

func TestInsertAllReplacesInsteadOfDuplicating(t *testing.T) {
	a, ads := loaded(t)
	_, err := a.InsertAll(context.Background(), ads) // importing again
	require.NoError(t, err)

	all, err := a.GetAllDocuments(context.Background())
	require.NoError(t, err)
	require.Len(t, all, len(ads))
}

// The index used to analyse titles with the standard analyzer, so these
// searches found nothing although the ads contain "apartament" and "центр".
func TestSearchMatchesRussianAndRomanianWordForms(t *testing.T) {
	a, _ := loaded(t)
	ctx := context.Background()

	for _, tc := range []struct{ stored, searched string }{
		{"apartament", "apartamente"}, // Romanian plural
		{"casa", "casele"},            // Romanian definite plural
		{"центр", "центра"},           // Russian genitive
		{"комнатная", "комнатные"},    // Russian plural adjective
	} {
		exact, err := a.FullTextSearch(ctx, tc.stored)
		require.NoError(t, err)
		require.NotEmpty(t, exact, "data should contain %q", tc.stored)

		inflected, err := a.FullTextSearch(ctx, tc.searched)
		require.NoError(t, err)
		require.NotEmpty(t, inflected, "searching %q should find ads that say %q", tc.searched, tc.stored)
	}
}

// The search text used to be spliced into the query JSON, so a quote broke
// the query or changed its structure.
func TestSearchTextIsNeverTreatedAsQuerySyntax(t *testing.T) {
	a, _ := loaded(t)
	for _, keyword := range []string{`casa" , "x":"`, `"}}, "size": 10000, "query": {"match_all": {}}}`, `\`, "teren\nvadul"} {
		_, err := a.FullTextSearch(context.Background(), keyword)
		require.NoError(t, err, "keyword %q", keyword)
	}
}

func TestInfiniteScrollPagesThroughEveryAdOnce(t *testing.T) {
	a, ads := loaded(t)
	seen := map[string]bool{}
	for from := 0; ; from += 20 {
		page, err := a.InfiniteScroll(context.Background(), from, 20)
		require.NoError(t, err)
		if len(page) == 0 {
			break
		}
		for _, ad := range page {
			require.False(t, seen[ad.ID], "ad %s appears on two pages", ad.ID)
			seen[ad.ID] = true
		}
	}
	require.Len(t, seen, len(ads))

	first, err := a.InfiniteScroll(context.Background(), 0, 3)
	require.NoError(t, err)
	require.GreaterOrEqual(t, first[0].Posted, first[1].Posted, "newest first")
}

func TestAggregateBySubcategoryCountsEveryAd(t *testing.T) {
	a, ads := loaded(t)
	want := map[string]int{}
	for _, ad := range ads {
		want[ad.Categories.Subcategory]++
	}
	got, err := a.AggregateBySubcategory(context.Background())
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestElasticsearchErrorsAreReturned(t *testing.T) {
	a := newTestAdapter(t)
	require.NoError(t, a.do(context.Background(), http.MethodDelete, "/"+a.cfg.Index, nil, nil))

	_, err := a.FullTextSearch(context.Background(), "casa")
	require.ErrorContains(t, err, "404")
}
