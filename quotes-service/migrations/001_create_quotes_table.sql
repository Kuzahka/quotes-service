CREATE TABLE IF NOT EXISTS quotes (
    id SERIAL PRIMARY KEY,
    author VARCHAR(100) NOT NULL CHECK (length(trim(author)) > 0),
    text TEXT NOT NULL CHECK (length(trim(text)) > 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Индексация для ускорения поиска
CREATE INDEX IF NOT EXISTS idx_quotes_author ON quotes (author);

CREATE INDEX IF NOT EXISTS idx_quotes_created_at ON quotes (created_at DESC);