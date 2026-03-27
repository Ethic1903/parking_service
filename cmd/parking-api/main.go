package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	v1 "parking-service/api/v1"
	"parking-service/internal/config"
	"parking-service/internal/parking"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if cfg.ConfigFile != "" {
		log.Printf("loaded config file: %s", cfg.ConfigFile)
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
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      mux,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("parking-api is listening on :%s (env=%s)", cfg.HTTP.Port, cfg.AppEnv)
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

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
