package extractor

import (
	"context"
	"io"
	"net/http"
)

type PDFExtractor struct {
	url    string
	client *http.Client
}

func NewPDFExtractor(url string) *PDFExtractor {
	return &PDFExtractor{
		url:    url + "/extract",
		client: &http.Client{},
	}
}

func (e *PDFExtractor) Extract(ctx context.Context, content io.Reader) (string, error) {
	return "", nil
}
