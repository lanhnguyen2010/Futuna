package main

import (
	"context"
	"log"

	"futuna/internal/analyzer"
	"futuna/internal/config"
	"futuna/internal/db"
	"futuna/internal/openai"
)

func main() {
	cfg := config.Load()
	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()

	llm := openai.New(cfg.OpenAIKey)
	svc := analyzer.NewService(database, llm)

	ctx := context.Background()
	if err := svc.AnalyzeAllAndStore(ctx); err != nil {
		log.Fatalf("analyze: %v", err)
	}
	log.Println("analysis completed")
}
