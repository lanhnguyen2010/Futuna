package models

import (
	"time"

	"github.com/jmoiron/sqlx/types"
)

// Analysis represents an analysis result for a ticker on a specific date.
type Analysis struct {
	ID         int64          `db:"id" json:"id"`
	Ticker     string         `db:"ticker" json:"ticker"`
	Date       time.Time      `db:"analyzed_at" json:"date"`
	ShortTerm  string         `db:"short_term" json:"short_term"`
	LongTerm   string         `db:"long_term" json:"long_term"`
	Strategies types.JSONText `db:"strategies" json:"strategies"`
	Overall    string         `db:"overall" json:"overall"`
	CreatedAt  time.Time      `db:"created_at" json:"created_at"`
}

// Ticker represents a HOSE ticker.
type Ticker struct {
	Symbol string `db:"symbol" json:"symbol"`
	Name   string `db:"name" json:"name"`
}
