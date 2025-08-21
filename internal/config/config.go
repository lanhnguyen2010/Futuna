package config

import (
	"log"
	"os"
	"strings"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	OpenAIKey      string
	DatabaseURL    string
	AnalyzeOnStart bool
}

// Load reads configuration from environment variables.
func Load() Config {
	cfg := Config{
		OpenAIKey:   os.Getenv("OPENAI_API_KEY"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		AnalyzeOnStart: func() bool {
			v := strings.ToLower(os.Getenv("ANALYZE_ON_START"))
			return v == "1" || v == "true"
		}(),
	}
	if cfg.OpenAIKey == "" {
		log.Println("warning: OPENAI_API_KEY is not set")
	}
	if cfg.DatabaseURL == "" {
		log.Println("warning: DATABASE_URL is not set")
	}
	return cfg
}
