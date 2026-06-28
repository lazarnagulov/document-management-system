package embedder

import (
	"context"
	"fmt"
	"net/http"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/httputil"
)

const (
	openAIEmbeddingURL = "https://api.openai.com/v1/embeddings"
	embeddingModel     = "text-embedding-3-small"
)

type Embedder interface {
	Embed(ctx context.Context, chunks []string) ([][]float32, error)
}

type embedder struct {
	apiKey string
	client *http.Client
}

func New(apiKey string) Embedder {
	return &embedder{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (e *embedder) Embed(ctx context.Context, chunks []string) ([][]float32, error) {
	if len(chunks) == 0 {
		return nil, nil
	}

	var result embeddingResponse
	err := httputil.PostJSON(
		ctx,
		e.client,
		openAIEmbeddingURL,
		embeddingRequest{Input: chunks, Model: embeddingModel},
		&result,
		map[string]string{"Authorization": "Bearer " + e.apiKey},
	)
	if err != nil {
		return nil, fmt.Errorf("openai embeddings: %w", err)
	}

	embeddings := make([][]float32, len(chunks))
	for _, d := range result.Data {
		if d.Index >= len(embeddings) {
			return nil, fmt.Errorf("unexpected embedding index %d", d.Index)
		}
		embeddings[d.Index] = d.Embedding
	}

	return embeddings, nil
}
