package extractor

import (
	"context"
	"io"
)

type TXTExtractor struct{}

func NewTXTExtractor() *TXTExtractor {
	return &TXTExtractor{}
}

func (e *TXTExtractor) Extract(ctx context.Context, content io.Reader) (string, error) {
	return "", nil
}
