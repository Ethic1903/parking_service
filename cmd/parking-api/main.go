package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	v1 "parking-service/internal/api/v1"
	"parking-service/internal/pkg/repository"
	"parking-service/internal/pkg/service"
	"parking-service/migrations"
	"parking-service/tools/config"
	"parking-service/tools/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if cfg.ConfigFile != "" {
		log.Printf("loaded config file: %s", cfg.ConfigFile)
	}

	db, driver, err := storage.OpenDB(cfg.Storage)
	if err != nil {
		log.Fatalf("failed to create DB client: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	if err := db.PingContext(ctx); err != nil {
		cancel()
		log.Fatalf("failed to ping DB: %v", err)
	}
	cancel()

	repo, err := repository.NewSQLRepository(db, driver)
	if err != nil {
		log.Fatalf("failed to build repository: %v", err)
	}

	initCtx, initCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	if err := migrations.Run(initCtx, db, driver); err != nil {
		initCancel()
		log.Fatalf("failed to run DB migrations: %v", err)
	}
	if err := repo.SeedSpots(initCtx, repository.DefaultSeedSpots()); err != nil {
		initCancel()
		log.Fatalf("failed to seed parking spots: %v", err)
	}
	initCancel()

	log.Printf("DB connected: driver=%s", driver)

	svc := service.NewService(repo)
	handler := v1.NewHandler(svc)

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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
