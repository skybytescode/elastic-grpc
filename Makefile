## test: run all tests (set ELASTICSEARCH_TEST_URL to include Elasticsearch)
test:
	@go test -race ./...

## up: build and start Elasticsearch and both services
up:
	@docker compose up --build

## proto: regenerate the gRPC code after changing proto/ingestworker.proto
proto:
	protoc -I proto --go_out=proto/ingestworker --go_opt=paths=source_relative \
		--go-grpc_out=proto/ingestworker --go-grpc_opt=paths=source_relative \
		proto/ingestworker.proto

.PHONY: test up proto
