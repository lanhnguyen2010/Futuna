-- +migrate Up
ALTER TABLE analyses
    ADD COLUMN IF NOT EXISTS overall_confidence INT,
    ADD COLUMN IF NOT EXISTS sources JSONB;

-- +migrate Down
ALTER TABLE analyses
    DROP COLUMN IF EXISTS overall_confidence,
    DROP COLUMN IF EXISTS sources;
