package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"

	"futuna/internal/analyzer"
	"futuna/internal/config"
	"futuna/internal/db"
	"futuna/internal/openai"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()

	llm := openai.New(cfg.OpenAIKey, cfg.OpenAIBaseURL, cfg.OpenAIModel)
	svc := analyzer.NewService(database, llm)

	ctx := context.Background()
	if err := svc.AnalyzeAllAndStore(ctx); err != nil {
		log.Fatalf("analyze: %v", err)
	}
	log.Println("analysis completed")
}
