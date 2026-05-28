package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"local/captcha-service/internal/config"
	captchahttp "local/captcha-service/internal/http"
	"local/captcha-service/internal/platform"
	"local/captcha-service/internal/store"
)

func main() {
	configPath := flag.String("c", "config.yml", "config file path")
	flag.Parse()

	cfg := config.Load(*configPath)

	st, closeStore, err := store.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("init store: %v", err)
	}
	defer closeStore()

	platformStore, err := platform.New(context.Background(), cfg)
	if err != nil {
		log.Fatalf("init platform store: %v", err)
	}
	defer platformStore.Close()

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           captchahttp.NewRouter(cfg, st, platformStore),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("captcha api listening on %s (store=%s platform=%s)", cfg.Addr, cfg.Store, cfg.PlatformStore)
		errCh <- srv.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		log.Printf("received %s, shutting down", sig)
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}