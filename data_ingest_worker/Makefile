build:
	@go build -o bin/data_ingest_worker ./cmd
run: build
	@./bin/data_ingest_worker
test :
	@go test -v ./...