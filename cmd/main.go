package main

import (
	"context"
	"net/http"
	"time"

	"subscription-service/internal/config"
	appdb "subscription-service/internal/db"
	apphttp "subscription-service/internal/http"
	"subscription-service/internal/logger"
	"subscription-service/internal/repository"
	"subscription-service/internal/service"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil { panic(err) }

	log, err := logger.New(cfg.Log.Level)
	if err != nil { panic(err) }
	defer log.Sync()

	ctx := context.Background()
	db, err := appdb.Connect(ctx, cfg.Postgres.DSN, cfg.Postgres.MinConns, cfg.Postgres.MaxConns)
	if err != nil { log.Fatal("db connect", zap.Error(err)) }
	defer db.Close()

	if err := appdb.RunMigrations(ctx, cfg.Postgres.DSN); err != nil { log.Fatal("migrations", zap.Error(err)) }

	repo := repository.NewSubscriptionRepository(db.Pool)
	svc := service.NewSubscriptionService(repo)

	h := apphttp.NewHandlers(log, svc)
	srv := apphttp.NewServer(log)
	srv.RegisterRoutes(h)

	server := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      srv.Router,
		ReadTimeout:  time.Duration(cfg.HTTP.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.HTTP.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cfg.HTTP.IdleTimeoutSeconds) * time.Second,
	}

	log.Info("starting http server", zap.String("addr", cfg.HTTP.Address))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("server", zap.Error(err))
	}
}


