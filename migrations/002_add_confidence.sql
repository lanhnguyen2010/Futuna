-- +migrate Up
ALTER TABLE analyses
    ADD COLUMN short_confidence INT,
    ADD COLUMN long_confidence INT;

-- +migrate Down
ALTER TABLE analyses
    DROP COLUMN IF EXISTS short_confidence,
    DROP COLUMN IF EXISTS long_confidence;
