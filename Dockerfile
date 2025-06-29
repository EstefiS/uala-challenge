FROM golang:1.24-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /app/main ./cmd/server/main.go

FROM debian:bookworm-slim

WORKDIR /

COPY --from=builder /app/main /main
RUN chmod +x /main
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

EXPOSE 8080
ENTRYPOINT ["/main"]