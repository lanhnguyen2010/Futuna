-- +migrate Up
CREATE TABLE IF NOT EXISTS tickers (
    symbol TEXT PRIMARY KEY,
    name   TEXT
);

CREATE TABLE IF NOT EXISTS analyses (
    id SERIAL PRIMARY KEY,
    ticker TEXT REFERENCES tickers(symbol),
    analyzed_at DATE NOT NULL,
    short_term TEXT,
    long_term TEXT,
    strategies JSONB,
    overall TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (ticker, analyzed_at)
);
ALTER TABLE analyses
    ADD COLUMN short_confidence INT,
    ADD COLUMN long_confidence INT;