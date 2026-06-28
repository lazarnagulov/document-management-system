CREATE TABLE document_chunks (
    id          UUID        PRIMARY KEY,
    document_id UUID        NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    content     TEXT        NOT NULL,
    embedding   vector(1536) NOT NULL,
    chunk_index INT         NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON document_chunks
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

CREATE INDEX idx_chunks_document_id ON document_chunks (document_id);