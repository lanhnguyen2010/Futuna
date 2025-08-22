package main

import (
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"

	"futuna/internal/analyzer"
	"futuna/internal/config"
	"futuna/internal/db"
	"futuna/internal/openai"
	"futuna/internal/web"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()
	llm := openai.New(cfg.OpenAIKey, cfg.OpenAIBaseURL, cfg.OpenAIModel)
	svc := analyzer.NewService(database, llm)
	if cfg.AnalyzeOnStart {
		ctx := context.Background()
		if err := svc.AnalyzeAllAndStore(ctx); err != nil {
			log.Fatalf("analyze: %v", err)
		}
		log.Println("initial analysis completed")
	}
	r := web.Router(svc)
	log.Println("web server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
