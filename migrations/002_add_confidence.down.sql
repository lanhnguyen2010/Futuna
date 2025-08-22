-- +migrate Down
ALTER TABLE analyses
    DROP COLUMN IF EXISTS short_confidence,
    DROP COLUMN IF EXISTS long_confidence;
