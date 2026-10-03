// Package db stores ads in Elasticsearch and searches them.
package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
)

// Config says where Elasticsearch is. Username and Password are optional.
type Config struct {
	URL      string
	Username string
	Password string
	Index    string // defaults to "adv"
}

type Adapter struct {
	cfg    Config
	client *http.Client
}

// NewAdapter connects to Elasticsearch, waiting up to wait for it to answer,
// and creates the index with its mapping if it does not exist yet.
func NewAdapter(ctx context.Context, cfg Config, wait time.Duration) (*Adapter, error) {
	if cfg.Index == "" {
		cfg.Index = "adv"
	}
	cfg.URL = strings.TrimRight(cfg.URL, "/")
	a := &Adapter{cfg: cfg, client: &http.Client{Timeout: 30 * time.Second}}

	deadline := time.Now().Add(wait)
	for {
		err := a.do(ctx, http.MethodGet, "/", nil, nil)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("elasticsearch at %s is not reachable: %w", cfg.URL, err)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	if err := a.createIndex(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

// indexDefinition analyses every title with both the Russian and the
// Romanian analyzer: ads are written in either language, often in the "wrong"
// field, so a search must match word forms ("квартиры" ~ "квартира",
// "apartamente" ~ "apartament") in both.
var indexDefinition = map[string]any{
	"mappings": map[string]any{
		"properties": map[string]any{
			"id": map[string]any{"type": "keyword"},
			"categories": map[string]any{"properties": map[string]any{
				"subcategory": map[string]any{"type": "keyword"},
			}},
			"title": map[string]any{"properties": map[string]any{
				"ro": bilingualText("romanian", "russian"),
				"ru": bilingualText("russian", "romanian"),
			}},
			"type":   map[string]any{"type": "keyword"},
			"posted": map[string]any{"type": "double"}, // Unix time in seconds
		},
	},
}

func bilingualText(primary, secondary string) map[string]any {
	return map[string]any{
		"type":     "text",
		"analyzer": primary,
		"fields":   map[string]any{secondary: map[string]any{"type": "text", "analyzer": secondary}},
	}
}

var searchFields = []string{"title.ro", "title.ro.russian", "title.ru", "title.ru.romanian"}

func (a *Adapter) createIndex(ctx context.Context) error {
	err := a.do(ctx, http.MethodHead, "/"+a.cfg.Index, nil, nil)
	if err == nil {
		return nil // already there
	}
	return a.do(ctx, http.MethodPut, "/"+a.cfg.Index, indexDefinition, nil)
}

// InsertAll indexes the ads in one bulk request, replacing ads with the same
// ID, and returns how many were indexed. It waits until they are searchable.
func (a *Adapter) InsertAll(ctx context.Context, ads []domain.Adv) (int, error) {
	if len(ads) == 0 {
		return 0, nil
	}
	var body bytes.Buffer
	enc := json.NewEncoder(&body)
	for _, ad := range ads {
		if ad.ID == "" {
			return 0, fmt.Errorf("ad without an id: %+v", ad)
		}
		if err := enc.Encode(map[string]any{"index": map[string]any{"_index": a.cfg.Index, "_id": ad.ID}}); err != nil {
			return 0, err
		}
		if err := enc.Encode(ad); err != nil {
			return 0, err
		}
	}

	var res struct {
		Errors bool `json:"errors"`
		Items  []map[string]struct {
			ID    string          `json:"_id"`
			Error json.RawMessage `json:"error"`
		} `json:"items"`
	}
	if err := a.send(ctx, http.MethodPost, "/_bulk?refresh=wait_for", "application/x-ndjson", &body, &res); err != nil {
		return 0, err
	}
	if res.Errors {
		for _, item := range res.Items {
			for _, r := range item {
				if len(r.Error) > 0 {
					return 0, fmt.Errorf("indexing ad %s failed: %s", r.ID, r.Error)
				}
			}
		}
	}
	return len(res.Items), nil
}

type searchResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source domain.Adv `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (r searchResponse) ads() []domain.Adv {
	ads := make([]domain.Adv, 0, len(r.Hits.Hits))
	for _, hit := range r.Hits.Hits {
		ads = append(ads, hit.Source)
	}
	return ads
}

// GetAllDocuments returns up to 1000 ads, newest first.
func (a *Adapter) GetAllDocuments(ctx context.Context) ([]domain.Adv, error) {
	return a.search(ctx, map[string]any{
		"size":  1000,
		"query": map[string]any{"match_all": map[string]any{}},
		"sort":  []any{map[string]any{"posted": "desc"}},
	})
}

// FullTextSearch finds ads whose title matches keyword in Russian or
// Romanian, best matches first.
func (a *Adapter) FullTextSearch(ctx context.Context, keyword string) ([]domain.Adv, error) {
	return a.search(ctx, map[string]any{
		"query": map[string]any{"multi_match": map[string]any{
			"query":  keyword, // encoded as JSON, never spliced into the query text
			"fields": searchFields,
			"type":   "most_fields",
		}},
	})
}

// InfiniteScroll returns size ads starting at offset from, newest first.
func (a *Adapter) InfiniteScroll(ctx context.Context, from, size int) ([]domain.Adv, error) {
	return a.search(ctx, map[string]any{
		"from":  from,
		"size":  size,
		"query": map[string]any{"match_all": map[string]any{}},
		"sort":  []any{map[string]any{"posted": "desc"}, map[string]any{"id": "asc"}},
	})
}

// AggregateBySubcategory counts ads per subcategory.
func (a *Adapter) AggregateBySubcategory(ctx context.Context) (map[string]int, error) {
	var res struct {
		Aggregations struct {
			Subcategories struct {
				Buckets []struct {
					Key      string `json:"key"`
					DocCount int    `json:"doc_count"`
				} `json:"buckets"`
			} `json:"subcategories"`
		} `json:"aggregations"`
	}
	query := map[string]any{
		"size": 0,
		"aggs": map[string]any{"subcategories": map[string]any{
			"terms": map[string]any{"field": "categories.subcategory", "size": 100},
		}},
	}
	if err := a.do(ctx, http.MethodPost, "/"+a.cfg.Index+"/_search", query, &res); err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, b := range res.Aggregations.Subcategories.Buckets {
		counts[b.Key] = b.DocCount
	}
	return counts, nil
}

func (a *Adapter) search(ctx context.Context, query map[string]any) ([]domain.Adv, error) {
	var res searchResponse
	if err := a.do(ctx, http.MethodPost, "/"+a.cfg.Index+"/_search", query, &res); err != nil {
		return nil, err
	}
	return res.ads(), nil
}

// do sends body as JSON and decodes the JSON answer into out.
func (a *Adapter) do(ctx context.Context, method, path string, body, out any) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}
	return a.send(ctx, method, path, "application/json", reader, out)
}

// send performs the request and turns any non-2xx answer into an error that
// carries Elasticsearch's own message.
func (a *Adapter) send(ctx context.Context, method, path, contentType string, body io.Reader, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, a.cfg.URL+path, body)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}
	if a.cfg.Username != "" {
		req.SetBasicAuth(a.cfg.Username, a.cfg.Password)
	}

	res, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("elasticsearch %s %s: %w", method, path, err)
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("elasticsearch %s %s: reading response: %w", method, path, err)
	}
	if res.StatusCode >= 300 {
		msg := strings.TrimSpace(string(data))
		if len(msg) > 500 {
			msg = msg[:500] + "…"
		}
		return fmt.Errorf("elasticsearch %s %s: %s: %s", method, path, res.Status, msg)
	}
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("elasticsearch %s %s: decoding response: %w", method, path, err)
		}
	}
	return nil
}
