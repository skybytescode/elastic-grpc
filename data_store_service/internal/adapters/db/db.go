package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/application/domain"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Adapter struct {
	db_cfg elasticsearch.Config
}

func NewAdapter() (*Adapter, error) {
	err := os.Setenv("ELASTIC_PASSWORD", "ELASTIC_PASSWORD")
	if err != nil {
		log.Fatalf("Error setting environment variable: %s", err)
	}

	// Get the Elasticsearch password from the environment variable
	password := os.Getenv("ELASTIC_PASSWORD")
	if password == "" {
		log.Fatal("ELASTIC_PASSWORD environment variable is not set")
	}
	db_cfg := elasticsearch.Config{
		Addresses: []string{
			//	"http://localhost:9200", // for local
			"http://elasticsearch:9200", // for docker
		},
		Username: "elastic",
		Password: password,
	}
	ad := &Adapter{db_cfg: db_cfg}
	ad.CheckHealth()
	err = ad.CreateIndex()
	if err != nil {
		fmt.Println("Already index created")
	}
	return ad, nil
}

type SearchHits struct {
	Hits struct {
		Hits []struct {
			Source domain.Adv `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (c *Adapter) CheckHealth() error {

	req, err := http.NewRequest("GET", c.db_cfg.Addresses[0], nil)
	if err != nil {
		return err
	}
	// Add basic authentication header
	req.SetBasicAuth(c.db_cfg.Username, c.db_cfg.Password)

	// Perform the request
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed check Elasticsearch health: %v", err)
	}

	log.Println("debug health check response: ", string(responseBody))

	return nil
}

func (c *Adapter) CreateIndex() error {
	body := `
	{
		"mappings": {
			"properties": {
				"id": {
					"type": "keyword"
				},
				"categories": {
					"properties": {
						"subcategory": {
							"type": "text",
							"fields": {
								"keyword": {
									"type": "keyword",
									"ignore_above": 256
								}
							}
						}
					}
				},
				"title": {
					"properties": {
						"ro": {
							"type": "text"
						},
						"ru": {
							"type": "text"
						}
					}
				},
				"type": {
					"type": "text"
				},
				"posted": {
					"type": "float"
				}
			}
		}
	}
	`

	req, err := http.NewRequest("PUT", c.db_cfg.Addresses[0]+"/adv", strings.NewReader(body))
	req.SetBasicAuth(c.db_cfg.Username, c.db_cfg.Password)
	if err != nil {
		return fmt.Errorf("failed to make a create index request, or already exists: %v", err)
	}

	httpClient := http.Client{}
	req.Header.Add("Content-type", "application/json")
	response, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make a httpServ call to create an index: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed read create index response: %v", err)
	}

	log.Println("debug create index response: ", string(responseBody))

	return nil
}

func (c *Adapter) InsertData(ctx context.Context, e domain.Adv) error {
	body, _ := json.Marshal(e)

	id := e.ID
	req, err := http.NewRequest("PUT", c.db_cfg.Addresses[0]+"/adv/_doc/"+id, bytes.NewBuffer(body))
	req.SetBasicAuth(c.db_cfg.Username, c.db_cfg.Password)
	if err != nil {
		return fmt.Errorf("failed to make a insert data request: %v", err)
	}

	httpClient := http.Client{}
	req.Header.Add("Content-type", "application/json")
	response, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make a httpServ call to insert data: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed read insert data response: %v", err)
	}

	log.Println("debug insert data response: ", string(responseBody))

	return nil
}

func (c *Adapter) SeedingData(ctx context.Context) error {
	if err := c.InsertData(ctx, domain.Adv{
		ID: fmt.Sprintf("38118540"),
		Categories: domain.Category{
			Subcategory: fmt.Sprintf("1401"),
		},
		Title: domain.Title{
			Ro: fmt.Sprintf("title_Ro"),
			Ru: fmt.Sprintf("title_Ru"),
		},
		Type:   "standard",
		Posted: 1486556302.101039,
	}); err != nil {
		return fmt.Errorf("failed seeding data: %v", err)
	}
	return nil
}

// Get all Documents
// 2.Поддерживает полнотекстовый поиск по полю title учитывая русскую и руммынсую морфологию.
func (a *Adapter) GetAllDocuments() ([]domain.Adv, error) {
	// Build the query for Elasticsearch with full-text search
	// for the example purpose I put a raw number.Default for _search is 10. Latter will be changed from parameters
	req, err := http.NewRequest("GET", a.db_cfg.Addresses[0]+"/adv/_search?size=1000", strings.NewReader(""))
	req.SetBasicAuth(a.db_cfg.Username, a.db_cfg.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to make a search data request: %v", err)
	}

	httpClient := http.Client{}
	req.Header.Add("Content-type", "application/json")
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make a httpServ call to search data: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read insert data response: %v", err)
	}

	var searchHits SearchHits
	if err := json.Unmarshal(responseBody, &searchHits); err != nil {
		return nil, fmt.Errorf("failed read unmarshal data response: %v", err)
	}

	var adv []domain.Adv
	for _, hit := range searchHits.Hits.Hits {
		adv = append(adv, hit.Source)
	}
	if len(adv) > 0 {
		fmt.Printf("Name is: %v", adv[0].Title)
	}
	return adv, nil
}

// 2.Поддерживает полнотекстовый поиск по полю title учитывая русскую и руммынсую морфологию.
func (a *Adapter) FullTextSearch(keyword string) ([]domain.Adv, error) {
	// Build the query for Elasticsearch with full-text search
	query := fmt.Sprintf(`
	{
		"query": {
			"multi_match": {
				"query": "%s",
				"fields": ["title.ro", "title.ru"],
				"type": "cross_fields"
			}
		}
	}`, keyword)

	req, err := http.NewRequest("GET", a.db_cfg.Addresses[0]+"/adv/_search", strings.NewReader(query))
	req.SetBasicAuth(a.db_cfg.Username, a.db_cfg.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to make a search data request: %v", err)
	}

	httpClient := http.Client{}
	req.Header.Add("Content-type", "application/json")
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make a httpServ call to search data: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read insert data response: %v", err)
	}

	var searchHits SearchHits
	if err := json.Unmarshal(responseBody, &searchHits); err != nil {
		return nil, fmt.Errorf("failed read unmarshal data response: %v", err)
	}

	var adv []domain.Adv
	for _, hit := range searchHits.Hits.Hits {
		adv = append(adv, hit.Source)
	}
	if len(adv) > 0 {
		fmt.Printf("Name is: %v", adv[0].Title)
	}
	return adv, nil
}

func (a *Adapter) InfiniteScroll(from int, size int) ([]domain.Adv, error) {
	// Build the query for Elasticsearch with pagination
	query := fmt.Sprintf(`
	{
		"from": %d,
		"size": %d,
		"query": {
			"match_all": {}
		}
	}`, from, size)

	req, err := http.NewRequest("GET", a.db_cfg.Addresses[0]+"/adv/_search", strings.NewReader(query))
	req.SetBasicAuth(a.db_cfg.Username, a.db_cfg.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to make a search data request: %v", err)
	}

	httpClient := http.Client{}
	req.Header.Add("Content-type", "application/json")
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make a httpServ call to search data: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read insert data response: %v", err)
	}

	var searchHits SearchHits
	if err := json.Unmarshal(responseBody, &searchHits); err != nil {
		return nil, fmt.Errorf("failed read unmarshal data response: %v", err)
	}

	var adv []domain.Adv
	for _, hit := range searchHits.Hits.Hits {
		adv = append(adv, hit.Source)
	}
	return adv, nil
}

type AggregationResponse struct {
	Aggregations struct {
		Subcategories struct {
			Buckets []struct {
				Key      string `json:"key"`
				DocCount int    `json:"doc_count"`
			} `json:"buckets"`
		} `json:"subcategories"`
	} `json:"aggregations"`
}

func (a *Adapter) AggregateBySubcategory() (map[string]int, error) {
	// Build the aggregation query for Elasticsearch
	query := `
	{
		"size": 0,
		"aggs": {
			"subcategories": {
				"terms": {
				"field": "categories.subcategory.keyword"
				}
			}
		}
	}`

	req, err := http.NewRequest("GET", a.db_cfg.Addresses[0]+"/adv/_search", strings.NewReader(query))
	req.SetBasicAuth(a.db_cfg.Username, a.db_cfg.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to make a search data request: %v", err)
	}

	httpClient := http.Client{}
	req.Header.Add("Content-type", "application/json")
	response, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make an HTTP call to search data: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read search data response: %v", err)
	}

	var aggResponse AggregationResponse
	if err := json.Unmarshal(responseBody, &aggResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search data response: %v", err)
	}

	subcategoryCounts := make(map[string]int)
	for _, bucket := range aggResponse.Aggregations.Subcategories.Buckets {
		subcategoryCounts[bucket.Key] = bucket.DocCount
	}

	return subcategoryCounts, nil
}
