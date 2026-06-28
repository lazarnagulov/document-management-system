package extractor

import (
	"context"
	"io"
)

type Config struct {
	URL string
}

type Extractor interface {
	Extract(ctx context.Context, content io.Reader) (string, error)
}
