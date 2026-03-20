package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	v1 "parking-service/api/v1"
	"parking-service/internal/parking"
)

func main() {
	port := strings.TrimSpace(os.Getenv("HTTP_PORT"))
	if port == "" {
		port = "8080"
	}

	repository := parking.NewInMemoryRepository([]parking.Spot{
		{ID: "A-101", Location: "center", VehicleType: "car", PricePerHour: 150, IsAvailable: true},
		{ID: "A-102", Location: "center", VehicleType: "car", PricePerHour: 170, IsAvailable: true},
		{ID: "B-201", Location: "airport", VehicleType: "car", PricePerHour: 220, IsAvailable: true},
		{ID: "C-301", Location: "station", VehicleType: "bike", PricePerHour: 80, IsAvailable: true},
	})
	service := parking.NewService(repository)
	handler := v1.NewHandler(service)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("parking-api is listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("received signal %s, shutting down", sig)
	case err := <-errCh:
		log.Fatalf("server failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
