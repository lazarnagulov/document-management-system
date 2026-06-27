package extractor

import (
	"context"
	"io"
)

type Extractor interface {
	Extract(ctx context.Context, content io.Reader) (string, error)
}
