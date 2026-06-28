package completer

import "github.com/google/uuid"

type SimilarChunk struct {
	ID         uuid.UUID
	Content    string
	Similarity float32
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}
