package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"

	"futuna/internal/models"
)

// OpenAI defines interface to get analysis.
type OpenAI interface {
	AnalyzeTickers(ctx context.Context) (string, string, string, error)
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

// AnalyzeAllAndStore runs analysis for all tickers and saves to DB.
func (s *Service) AnalyzeAllAndStore(ctx context.Context) error {
	reqJSON, respJSON, result, err := s.llm.AnalyzeTickers(ctx)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO openai_logs (request, response) VALUES ($1,$2)`, types.JSONText(reqJSON), types.JSONText(respJSON)); err != nil {
		return err
	}
	var payload struct {
		AsOf    string `json:"as_of"`
		Tickers []struct {
			Ticker    string `json:"ticker"`
			ShortTerm struct {
				Recommendation string `json:"recommendation"`
				Confidence     int    `json:"confidence"`
				Reason         string `json:"reason"`
			} `json:"short_term"`
			LongTerm struct {
				Recommendation string `json:"recommendation"`
				Confidence     int    `json:"confidence"`
				Reason         string `json:"reason"`
			} `json:"long_term"`
			Strategies []struct {
				Name   string `json:"name"`
				Stance string `json:"stance"`
				Note   string `json:"note"`
			} `json:"strategies"`
			Overall struct {
				Recommendation string `json:"recommendation"`
				Confidence     int    `json:"confidence"`
				Reason         string `json:"reason"`
			} `json:"overall"`
		} `json:"tickers"`
		Sources []string `json:"sources"`
	}
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		return err
	}
	date, err := time.Parse(time.RFC3339, payload.AsOf)
	if err != nil {
		date = time.Now()
	}
	date = date.Truncate(24 * time.Hour)
	sourcesJSON, _ := json.Marshal(payload.Sources)
	for _, item := range payload.Tickers {
		if _, err := s.db.ExecContext(ctx, `INSERT INTO tickers (symbol) VALUES ($1) ON CONFLICT DO NOTHING`, item.Ticker); err != nil {
			return err
		}
		short := fmt.Sprintf("%s - %s", item.ShortTerm.Recommendation, item.ShortTerm.Reason)
		shortConf := item.ShortTerm.Confidence
		long := fmt.Sprintf("%s - %s", item.LongTerm.Recommendation, item.LongTerm.Reason)
		longConf := item.LongTerm.Confidence
		overall := fmt.Sprintf("%s - %s", item.Overall.Recommendation, item.Overall.Reason)
		overallConf := item.Overall.Confidence
		strategiesJSON, _ := json.Marshal(item.Strategies)

		_, err := s.db.ExecContext(ctx, `INSERT INTO analyses (ticker, analyzed_at, short_term, short_confidence, long_term, long_confidence, strategies, overall, overall_confidence, sources, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW()) ON CONFLICT (ticker, analyzed_at) DO UPDATE SET short_term=EXCLUDED.short_term, short_confidence=EXCLUDED.short_confidence, long_term=EXCLUDED.long_term, long_confidence=EXCLUDED.long_confidence, strategies=EXCLUDED.strategies, overall=EXCLUDED.overall, overall_confidence=EXCLUDED.overall_confidence, sources=EXCLUDED.sources`,
			item.Ticker, date, short, shortConf, long, longConf, strategiesJSON, overall, overallConf, sourcesJSON)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListAnalyses returns analyses for a given date.
func (s *Service) ListAnalyses(ctx context.Context, date time.Time) ([]models.Analysis, error) {
	rows := []models.Analysis{}
	err := s.db.SelectContext(ctx, &rows, `SELECT id, ticker, analyzed_at, short_term, short_confidence, long_term, long_confidence, strategies, overall, overall_confidence, sources, created_at FROM analyses WHERE analyzed_at=$1 ORDER BY ticker`, date)
	return rows, err
}

// ListTickers returns all tickers.
func (s *Service) ListTickers(ctx context.Context) ([]models.Ticker, error) {
	rows := []models.Ticker{}
	err := s.db.SelectContext(ctx, &rows, `SELECT symbol, name FROM tickers ORDER BY symbol`)
	return rows, err
}
