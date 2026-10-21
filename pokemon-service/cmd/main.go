package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pokemon-service/config"
	"pokemon-service/internal/delivery/http/router"
	postgresRepo "pokemon-service/internal/repository/postgres"
	"pokemon-service/internal/usecase"
	"pokemon-service/pkg/logger"
)

// @title			Pokemon Service API
// @version		1.0
// @description	Boilerplate REST API CRUD Pokémon (Go + Gin + GORM + PostgreSQL, Clean Architecture).
// @BasePath		/api/v1
// @schemes		http https
func main() {
	cfg, err := config.Load()
	if err != nil {
		logger.New("error").Error("load config", "err", err)
		os.Exit(1)
	}
	log := logger.New(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := postgresRepo.NewDB(ctx, cfg.Database)
	if err != nil {
		log.Error("init database", "err", err)
		os.Exit(1)
	}

	// Dependency injection manual: repository → use case → delivery
	pokemonRepo := postgresRepo.NewPokemonRepository(db)
	pokemonUC := usecase.NewPokemonUseCase(pokemonRepo)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router.New(cfg, log, pokemonUC, postgresRepo.NewHealthChecker(db)),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("HTTP server started", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", "err", err)
	}

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	log.Info("server stopped")
}
