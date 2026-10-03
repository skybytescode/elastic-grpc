package httpServ

import (
	"context"
	"fmt"
	"github.com/skybytescode/elastic-grpc/data_store_service/internal/ports"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Adapter struct {
	api    ports.APIPort
	port   int
	server *http.Server
}

func NewAdapter(api ports.APIPort, port int) *Adapter {
	return &Adapter{api: api, port: port}
}

// routes maps the URLs to their handlers.
func (a *Adapter) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/create", a.createItemHandler)     // POST: ingest worker -> Elasticsearch
	mux.HandleFunc("/getalldocs", a.getAllDocsHandler) // all ads, newest first
	mux.HandleFunc("/searchTitle", a.searchByTitle)    // ?title=apartamente
	mux.HandleFunc("/scroll", a.scrollSearch)          // ?from=0&size=10
	mux.HandleFunc("/aggsub", a.aggSubcategory)        // ad count per subcategory
	return mux
}

func (a *Adapter) Run() {
	a.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", a.port),
		Handler:           a.routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Start the server in a separate goroutine
	go func() {
		log.Printf("Server is running on port %d\n", a.port)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on port %d: %v\n", a.port, err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
