package extractor

import (
	"context"
	"io"
)

type PDFExtractor struct {
	remoteExtractor
}

func NewPDFExtractor(url string) *PDFExtractor {
	return &PDFExtractor{newRemoteExtractor(url+"/extract", "document.pdf", "application/pdf")}
}

func (e *PDFExtractor) Extract(ctx context.Context, content io.Reader) (string, error) {
	return e.extract(ctx, content)
}
