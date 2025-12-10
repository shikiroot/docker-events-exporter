# Docker Events Exporter

A Prometheus Exporter for monitoring Docker container exit codes. It listens on port 8000 by default.

## Features

- Listens for Docker `die` events
- Exposes `docker_events` counter metric
- Labels included: `container_image`, `container_name`, `exit_code`
- Supports Graceful Shutdown

## Build Instructions

### Method 1: Using Makefile (Recommended)

This project includes a Makefile for easy building:

```bash
# 1. Build for Linux amd64 (Target Environment)
make build-linux
# Output: docker-events-exporter-linux-amd64

# 2. Build for Local Environment (Mac/Windows)
make build
# Output: docker-events-exporter

# 3. Build Docker Image
make docker-build
```

### Method 2: Using Go Directly

```bash
# Build as Linux amd64 binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o docker-events-exporter main.go
```

## How to Run

### Option 1: Run as System Service (Systemd)

If you need to run this as a background service on a Linux server, use the provided Systemd configuration.

1. **Install Binary**:
   Copy the compiled `docker-events-exporter-linux-amd64` to `/usr/local/bin/docker-events-exporter`:

   ```bash
   cp docker-events-exporter-linux-amd64 /usr/local/bin/docker-events-exporter
   chmod +x /usr/local/bin/docker-events-exporter
   ```

2. **Install Service File**:
   Copy `docker-events-exporter.service` to `/etc/systemd/system/`:

   ```bash
   sudo cp docker-events-exporter.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable docker-events-exporter
   sudo systemctl start docker-events-exporter
   ```

3. **Check Status**:
   ```bash
   systemctl status docker-events-exporter
   ```

### Option 2: Run with Docker

```bash
docker run -d \
  --name docker-events-exporter \
  -p 8000:8000 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  docker-events-exporter
```

## Metrics Example

Visit `http://localhost:8000/metrics`:

```text
# HELP docker_events Count of docker container exit events
# TYPE docker_events counter
docker_events{container_image="nginx:latest",container_name="my-nginx",exit_code="0"} 1
docker_events{container_image="alpine:3.14",container_name="job-runner",exit_code="137"} 3
```
