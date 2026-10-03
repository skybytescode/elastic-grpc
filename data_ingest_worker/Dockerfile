# Built from the repository root: docker build -f data_ingest_worker/Dockerfile .
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags='-s' -o /out/data_ingest_worker ./data_ingest_worker/cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/data_ingest_worker /data_ingest_worker
COPY data_ingest_worker/adapters/db/data.json /data.json
EXPOSE 4001
ENTRYPOINT ["/data_ingest_worker"]
