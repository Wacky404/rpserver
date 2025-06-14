# Stage 1: Builder
FROM golang:1.21 AS builder

WORKDIR /rpserver

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o bin/rpserver ./cmd/rpserver/main.go

# Stage 2: Runtime
FROM debian:bookworm-slim

WORKDIR /rpserver

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /rpserver/bin/rpserver .
# Do not copy certs, they’ll be mounted

EXPOSE 8080

CMD ["./rpserver", "--cert", "certs/localhost.pem", "--key", "certs/localhost-key.pem"]
