package extractor

import (
	"context"
	"io"
)

type DOCXExtractor struct {
}

func NewDOCXExtractor() *DOCXExtractor {
	return &DOCXExtractor{}
}

func (e *DOCXExtractor) Extract(_ context.Context, content io.Reader) (string, error) {
	return "", nil
}
