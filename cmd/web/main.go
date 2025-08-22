package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"futuna/internal/analyzer"
	"futuna/internal/config"
	"futuna/internal/db"
	"futuna/internal/openai"
	"futuna/internal/web"
)

func main() {
	// Load .env if present (non-fatal if missing)
	_ = godotenv.Load()

	// App config
	cfg := config.Load()

	// DB connect
	database := db.Connect(cfg.DatabaseURL)
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("db close: %v", err)
		}
	}()

	// LLM client (uses your fixed internal/openai wrapper)
	llm := openai.New(cfg.OpenAIKey, cfg.OpenAIBaseURL, cfg.OpenAIModel)

	// Service layer
	svc := analyzer.NewService(database, llm)

	// Optional boot-time analysis
	if cfg.AnalyzeOnStart {
		log.Println("starting initial analysis …")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		if err := svc.AnalyzeAllAndStore(ctx); err != nil {
			log.Fatalf("analyze: %v", err)
		}
		log.Println("initial analysis completed")
	}

	// Router (from your internal/web)
	handler := web.Router(svc)

	// Port: prefer CFG or PORT env; fallback 8080
	addr := ":" + pickPort(cfg.Port)

	// HTTP server with sane timeouts
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		// Listen for SIGINT/SIGTERM
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("signal received: %s — shutting down …", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Printf("web server starting on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
	<-idleConnsClosed
	log.Println("server stopped cleanly")
}

// pickPort returns a port string (without leading colon)
// from cfg value, then $PORT, else "8080".
func pickPort(cfgPort int) string {
	if cfgPort > 0 {
		return strconv.Itoa(cfgPort)
	}
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}
