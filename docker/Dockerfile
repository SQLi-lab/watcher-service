FROM golang:1.23 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod tidy

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o watcher cmd/watcher/watcher.go


FROM debian:bullseye-slim

RUN apt-get update && apt-get install -y \
    docker.io \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /build/watcher .

ENTRYPOINT ["/app/watcher"]
