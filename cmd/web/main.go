package main

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"futuna/internal/analyzer"
	"futuna/internal/config"
	"futuna/internal/db"
	"futuna/internal/openai"
	"futuna/internal/web"
)

func main() {
	cfg := config.Load()
	database := db.Connect(cfg.DatabaseURL)
	defer database.Close()
	llm := openai.New(cfg.OpenAIKey)
	svc := analyzer.NewService(database, llm)
	if cfg.AnalyzeOnStart {
		ctx := context.Background()
		tickers, err := svc.ListTickers(ctx)
		if err != nil {
			log.Fatalf("load tickers: %v", err)
		}
		date := time.Now().Truncate(24 * time.Hour)
		var wg sync.WaitGroup
		for _, t := range tickers {
			wg.Add(1)
			go func(ticker string) {
				defer wg.Done()
				if err := svc.AnalyzeAndStore(ctx, ticker, date); err != nil {
					log.Printf("analyze %s: %v", ticker, err)
				}
			}(t.Symbol)
		}
		wg.Wait()
		log.Println("initial analysis completed")
	}
	r := web.Router(svc)
	log.Println("web server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
