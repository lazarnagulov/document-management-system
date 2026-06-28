package extractor

import (
	"context"
	"io"
)

type Config struct {
	PDFURL  string
	DOCXURL string
}

type Extractor interface {
	Extract(ctx context.Context, content io.Reader) (string, error)
}
