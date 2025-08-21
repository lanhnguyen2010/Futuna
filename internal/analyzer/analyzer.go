package analyzer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"

	"futuna/internal/models"
)

// OpenAI defines interface to get analysis.
type OpenAI interface {
	AnalyzeVN30(ctx context.Context) (string, error)
}

// Service orchestrates OpenAI calls and persistence.
type Service struct {
	db  *sqlx.DB
	llm OpenAI
}

// NewService creates a new Service.
func NewService(db *sqlx.DB, llm OpenAI) *Service {
	return &Service{db: db, llm: llm}
}

// AnalyzeAllAndStore runs analysis for all VN30 tickers and saves to DB.
func (s *Service) AnalyzeAllAndStore(ctx context.Context, date time.Time) error {
	result, err := s.llm.AnalyzeVN30(ctx)
	if err != nil {
		return err
	}
	var payload map[string]struct {
		ShortTerm  string `json:"short_term"`
		LongTerm   string `json:"long_term"`
		Strategies []struct {
			Name   string `json:"name"`
			Stance string `json:"stance"`
			Note   string `json:"note"`
		} `json:"strategies"`
	}
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		return err
	}
	for ticker, p := range payload {
		strategiesJSON, _ := json.Marshal(p.Strategies)
		_, err := s.db.ExecContext(ctx, `INSERT INTO analyses (ticker, analyzed_at, short_term, long_term, strategies, created_at, overall)
 VALUES ($1,$2,$3,$4,$5,NOW(),'')
 ON CONFLICT (ticker, analyzed_at) DO UPDATE SET short_term=EXCLUDED.short_term, long_term=EXCLUDED.long_term, strategies=EXCLUDED.strategies`,
			ticker, date, p.ShortTerm, p.LongTerm, strategiesJSON)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListAnalyses returns analyses for a given date.
func (s *Service) ListAnalyses(ctx context.Context, date time.Time) ([]models.Analysis, error) {
	rows := []models.Analysis{}
	err := s.db.SelectContext(ctx, &rows, `SELECT id, ticker, analyzed_at, short_term, long_term, strategies, created_at, overall FROM analyses WHERE analyzed_at=$1 ORDER BY ticker`, date)
	return rows, err
}

// ListTickers returns all tickers.
func (s *Service) ListTickers(ctx context.Context) ([]models.Ticker, error) {
	rows := []models.Ticker{}
	err := s.db.SelectContext(ctx, &rows, `SELECT symbol, name FROM tickers ORDER BY symbol`)
	return rows, err
}
