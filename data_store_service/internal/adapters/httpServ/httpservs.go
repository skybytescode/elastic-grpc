package httpServ

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const maxPageSize = 100

func writeJSON(w http.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("writing response: %v", err)
	}
}

// writeError reports a failure. Problems with the request are the client's
// (400); anything else comes from Elasticsearch or the ingest worker (502).
func writeError(w http.ResponseWriter, status int, err error) {
	log.Printf("request failed: %v", err)
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func badRequest(w http.ResponseWriter, message string) {
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": message})
}

// createItemHandler copies the ingest worker's ads into Elasticsearch.
func (a *Adapter) createItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "use POST"})
		return
	}
	n, err := a.api.PlaceData(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"indexed": n})
}

func (a *Adapter) getAllDocsHandler(w http.ResponseWriter, r *http.Request) {
	response, err := a.api.GetAllData(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *Adapter) searchByTitle(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		badRequest(w, "the title parameter is required")
		return
	}
	response, err := a.api.TextSearch(r.Context(), title)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *Adapter) scrollSearch(w http.ResponseWriter, r *http.Request) {
	from, err := intParam(r, "from", 0)
	if err != nil || from < 0 {
		badRequest(w, "from must be a number ≥ 0")
		return
	}
	size, err := intParam(r, "size", 10)
	if err != nil || size < 1 || size > maxPageSize {
		badRequest(w, "size must be a number from 1 to 100")
		return
	}
	response, err := a.api.ScrollSearch(r.Context(), from, size)
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *Adapter) aggSubcategory(w http.ResponseWriter, r *http.Request) {
	response, err := a.api.AggregateSubcategory(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func intParam(r *http.Request, key string, fallback int) (int, error) {
	v := r.FormValue(key)
	if v == "" {
		return fallback, nil
	}
	return strconv.Atoi(v)
}
