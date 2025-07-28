package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tacacs-zitadel-server/config"
	"tacacs-zitadel-server/handlers"
	"tacacs-zitadel-server/tacacs_tacquito"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.Load()
	
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)
	logger.SetFormatter(&logrus.JSONFormatter{})

	logger.Info("Starting TACACS+ server with Zitadel integration")

	tacacsServer, err := tacacs_tacquito.NewTacacsServer(cfg, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create TACACS+ server")
	}

	router := mux.NewRouter()
	router.HandleFunc("/health", handlers.HealthHandler).Methods("GET")
	router.HandleFunc("/metrics", handlers.MetricsHandler).Methods("GET")

	httpServer := &http.Server{
		Addr:    cfg.HTTPListenAddress,
		Handler: router,
	}

	go func() {
		logger.WithField("address", cfg.HTTPListenAddress).Info("Starting HTTP server")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("HTTP server failed")
		}
	}()

	go func() {
		logger.WithField("address", cfg.TACACSListenAddress).Info("Starting TACACS+ server")
		if err := tacacsServer.Start(); err != nil {
			logger.WithError(err).Fatal("TACACS+ server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("HTTP server shutdown error")
	}

	if err := tacacsServer.Stop(); err != nil {
		logger.WithError(err).Error("TACACS+ server shutdown error")
	}

	logger.Info("Servers shut down successfully")
}