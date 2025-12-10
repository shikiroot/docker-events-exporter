FROM golang:1.21 AS builder

WORKDIR /app

COPY go.mod ./
# COPY go.sum ./ 
# Since we generated the files manually, we run tidy to ensure deps are fetched
RUN go mod tidy

COPY . .

# Build for Linux 64-bit (amd64)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o docker-events-exporter main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/docker-events-exporter .

# Expose the metrics port
EXPOSE 8000

# Run the binary
ENTRYPOINT ["./docker-events-exporter"]
