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

func (a *Adapter) Run() {
	mux := http.NewServeMux()

	// Define your routes here
	mux.HandleFunc("/seeddata", a.seedItemHandler)     // http://localhost:8080/seeddata
	mux.HandleFunc("/create", a.createItemHandler)     // http://localhost:8080/create
	mux.HandleFunc("/getalldocs", a.getAllDocsHandler) // http://localhost:8080/getalldocs
	mux.HandleFunc("/searchTitle", a.searchByTitle)    //http://localhost:8080/searchTitle?title=se
	mux.HandleFunc("/scroll", a.scrollSearch)          // http://localhost:8080/scroll?from=1&size=5
	mux.HandleFunc("/aggsub", a.aggSubcategory)        // http://localhost:8080/aggsub
	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", a.port),
		Handler: mux,
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
