-- +migrate Down
ALTER TABLE analyses
    DROP COLUMN IF EXISTS overall_confidence,
    DROP COLUMN IF EXISTS sources;
