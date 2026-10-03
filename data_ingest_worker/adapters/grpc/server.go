package grpc

import (
	"fmt"
	"github.com/skybytescode/elastic-grpc/data_ingest_worker/config"
	"github.com/skybytescode/elastic-grpc/data_ingest_worker/internal/ports"
	"github.com/skybytescode/elastic-grpc/proto/ingestworker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

type Adapter struct {
	api    ports.APIPort
	port   int
	server *grpc.Server
	ingestworker.UnimplementedRetrieveDataServer
}

func NewAdapter(api ports.APIPort, port int) *Adapter {
	return &Adapter{api: api, port: port}
}

// Run listens on the configured port and serves until the process stops.
func (a *Adapter) Run() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		log.Fatalf("failed to listen on port %d, error: %v", a.port, err)
	}
	log.Printf("starting grpc worker service on port %d ...", a.port)
	if err := a.Serve(listen); err != nil {
		log.Fatalf("failed to serve grpc on port %d: %v", a.port, err)
	}
}

// Serve serves gRPC on lis until Stop is called.
func (a *Adapter) Serve(lis net.Listener) error {
	a.server = grpc.NewServer()
	ingestworker.RegisterRetrieveDataServer(a.server, a)
	if config.GetEnv() == "development" {
		reflection.Register(a.server)
	}
	return a.server.Serve(lis)
}

func (a *Adapter) Stop() {
	a.server.Stop()
}
