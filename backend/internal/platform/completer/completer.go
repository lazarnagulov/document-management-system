package completer

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/httputil"
)

const (
	openAIChatURL = "https://api.openai.com/v1/chat/completions"
	chatModel     = "gpt-4o-mini"
)

type Completer interface {
	Complete(ctx context.Context, question string, chunks []*SimilarChunk) (string, error)
}

type completer struct {
	apiKey string
	client *http.Client
}

func New(apiKey string) Completer {
	return &completer{apiKey: apiKey, client: &http.Client{}}
}

func (c *completer) Complete(ctx context.Context, question string, chunks []*SimilarChunk) (string, error) {
	context := buildContext(chunks)
	payload := chatRequest{
		Model: chatModel,
		Messages: []message{
			{
				Role: "system",
				Content: `You are a helpful assistant that answers questions based strictly on the provided document context.
If the answer cannot be found in the context, say so clearly.
Do not make up information.`,
			},
			{
				Role: "user",
				Content: fmt.Sprintf(
					"Context from document:\n\n%s\n\nQuestion: %s",
					context, question,
				),
			},
		},
	}

	var result chatResponse
	err := httputil.PostJSON(
		ctx,
		c.client,
		openAIChatURL,
		payload,
		&result,
		map[string]string{"Authorization": "Bearer " + c.apiKey},
	)
	if err != nil {
		return "", fmt.Errorf("openai chat: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from openai")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

func buildContext(chunks []*SimilarChunk) string {
	var sb strings.Builder
	for i, chunk := range chunks {
		if i > 0 {
			sb.WriteString("\n\n---\n\n")
		}
		sb.WriteString(chunk.Content)
	}
	return sb.String()
}
