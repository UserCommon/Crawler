CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS pages (
    id BIGSERIAL PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,
    html TEXT,

    -- Результаты анализа от LLM
    structure_analysis JSONB DEFAULT '{}'::jsonb,
    content_analysis JSONB DEFAULT '{}'::jsonb,

    -- Векторное представление для RAG (768 - для модели nomic-embed-text)
    embedding vector(384),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pages_url ON pages(url);

CREATE INDEX IF NOT EXISTS idx_pages_structure ON pages USING GIN (structure_analysis);
CREATE INDEX IF NOT EXISTS idx_pages_content ON pages USING GIN (content_analysis);

CREATE INDEX IF NOT EXISTS idx_pages_embedding ON pages
USING hnsw (embedding vector_cosine_ops);
