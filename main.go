package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/app"
	"github.com/AdrianJanczenia/adrianjanczenia.dev_gateway-service/internal/registry"
)

func main() {
	cfg, err := registry.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: could not load configuration: %v", err)
	}
	application, err := app.Build(cfg)
	if err != nil {
		log.Fatalf("FATAL: could not build application: %v", err)
	}

	serverErrors := make(chan error, 1)
	go func() {
		if err := application.RunHTTP(); !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, syscall.SIGINT, syscall.SIGTERM)
	failed := false
	select {
	case sig := <-shutdownChannel:
		log.Printf("INFO: received signal %s. Shutting down...", sig)
	case err := <-serverErrors:
		log.Printf("ERROR: HTTP server failed: %v", err)
		failed = true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	application.Shutdown(ctx)
	log.Println("INFO: application shutdown complete.")
	if failed {
		os.Exit(1)
	}
}
