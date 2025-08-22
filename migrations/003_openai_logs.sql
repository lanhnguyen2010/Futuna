CREATE TABLE IF NOT EXISTS openai_logs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    request JSONB NOT NULL,
    response JSONB NOT NULL
);
