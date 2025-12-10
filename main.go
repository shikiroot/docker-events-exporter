package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metric definition
var (
	dockerEventsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "docker_events",
			Help: "Count of docker container exit events",
		},
		[]string{"container_image", "container_name", "exit_code"},
	)
)

func main() {
	// Register Prometheus metrics
	prometheus.MustRegister(dockerEventsCounter)

	// Create a context that is cancelled on OS signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the event listener in a background goroutine
	go startEventListener(ctx)

	// Start HTTP server in a background goroutine
	srv := &http.Server{Addr: ":8000"}
	go func() {
		log.Println("Starting exporter on :8000/metrics")
		http.Handle("/metrics", promhttp.Handler())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting HTTP server: %v", err)
		}
	}()

	// Wait for signal
	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	// Shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	log.Println("Exporter stopped")
}

func startEventListener(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := listenDockerEvents(ctx)
			if err != nil {
				log.Printf("Error in event listener: %v. Reconnecting in 5 seconds...", err)
				select {
				case <-time.After(5 * time.Second):
					continue
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func listenDockerEvents(ctx context.Context) error {
	// Initialize Docker client
	// WithAPIVersionNegotiation ensures compatibility with the server version
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	// Set up filters to listen only for container 'die' events
	filter := filters.NewArgs()
	filter.Add("type", "container")
	filter.Add("event", "die")

	// Verify connection
	_, err = cli.Ping(ctx)
	if err != nil {
		return err
	}

	// Start listening for events
	msgs, errs := cli.Events(ctx, types.EventsOptions{
		Filters: filter,
	})

	log.Println("Successfully connected to Docker Daemon. Waiting for container exit (die) events...")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errs:
			if err != nil {
				return err
			}
		case msg := <-msgs:
			processEvent(msg)
		}
	}
}

func processEvent(msg events.Message) {
	// Extract basic information
	// The actor attributes contain the details we need
	attrs := msg.Actor.Attributes

	// Get container name
	containerName := attrs["name"]
	if containerName == "" {
		// Fallback if name is empty, though unlikely for container events
		containerName = msg.Actor.ID[:12]
	}

	// Get container image
	containerImage := attrs["image"]
	if containerImage == "" {
		containerImage = msg.From
	}

	// Get exit code
	// For 'die' events, exitCode is usually present in attributes
	exitCode := attrs["exitCode"]
	if exitCode == "" {
		exitCode = "unknown"
	}

	log.Printf("Event: Container='%s' Image='%s' ExitCode='%s'", containerName, containerImage, exitCode)

	// Increment the counter
	dockerEventsCounter.WithLabelValues(containerImage, containerName, exitCode).Inc()
}
