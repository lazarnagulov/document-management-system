CREATE OR REPLACE FUNCTION match_chunks(
    query_embedding vector(1536),
    match_document_id UUID,
    match_count INT DEFAULT 5
)
RETURNS TABLE (
    id UUID,
    content TEXT,
    similarity FLOAT
)
LANGUAGE SQL STABLE
AS $$
    SELECT
        id,
        content,
        1 - (embedding <=> query_embedding) AS similarity
    FROM document_chunks
    WHERE document_id = match_document_id
    ORDER BY embedding <=> query_embedding
    LIMIT match_count;
$$;