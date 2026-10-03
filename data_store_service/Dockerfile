# Built from the repository root: docker build -f data_store_service/Dockerfile .
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags='-s' -o /out/data_store_service ./data_store_service/cmd

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/data_store_service /data_store_service
EXPOSE 8080
ENTRYPOINT ["/data_store_service"]
