-- +migrate Up
ALTER TABLE analyses
    ADD COLUMN IF NOT EXISTS overall_confidence INT,
    ADD COLUMN IF NOT EXISTS sources JSONB;
