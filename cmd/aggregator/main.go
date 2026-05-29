package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/config"
	"ftso-prices/internal/registry"
	"ftso-prices/internal/server"
	"ftso-prices/internal/service"
)

func main() {
	cfg := config.Load(envPath())
	log.Printf("Starting aggregator: %d assets, %d quotes, port %d, timeout %s",
		len(cfg.Assets), len(cfg.Quotes), cfg.Port, cfg.Timeout)

	srv := server.New()
	svc := service.New(srv)
	ctx := client.NewContext(cfg)

	// Base-asset filter shared by all producers, then the service-level quote filter.
	push := client.FilterPush(cfg.AssetSet, svc.Push)

	mux := http.NewServeMux()
	mux.HandleFunc("/ticker", srv.Handler)
	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: mux}

	stop := make(chan struct{})
	var wg sync.WaitGroup

	// REST resumes as a fallback if the websocket feed goes silent this long.
	const restFallbackAfter = 10 * time.Second

	hybridExchanges := registry.EnabledHybrid(cfg.Enabled)
	for _, h := range hybridExchanges {
		runner := client.NewHybridRunner(h.WS, h.REST, ctx, push, restFallbackAfter)
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner.Run(stop)
		}()
	}

	wsExchanges := registry.EnabledWs(cfg.Enabled)
	for _, ex := range wsExchanges {
		runner := client.NewWsRunner(ex, ctx, push)
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner.Run(stop)
		}()
	}

	restExchanges := registry.EnabledRest(cfg.Enabled)
	for _, ex := range restExchanges {
		runner := client.NewRestRunner(ex, ctx, push)
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner.Run(stop)
		}()
	}

	log.Printf("Enabled exchanges: %d hybrid (WS+REST), %d websocket, %d REST", len(hybridExchanges), len(wsExchanges), len(restExchanges))

	go func() {
		log.Printf("Ticker websocket server listening on :%d/ticker", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	log.Printf("Shutting down ...")
	close(stop)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
	wg.Wait()
}

// envPath looks for a .env file next to the binary's project (../../.env from
// the go module root) so the existing project .env is reused.
func envPath() string {
	for _, p := range []string{".env", "../.env", "../../.env"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ".env"
}
