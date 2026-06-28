package extractor

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/httputil"
)

type remoteExtractor struct {
	url         string
	filename    string
	contentType string
	client      *http.Client
}

func newRemoteExtractor(url, filename, contentType string) remoteExtractor {
	return remoteExtractor{
		url:         url,
		filename:    filename,
		contentType: contentType,
		client:      &http.Client{},
	}
}

func (r *remoteExtractor) extract(ctx context.Context, content io.Reader) (string, error) {
	var result struct {
		Text string `json:"text"`
	}
	if err := httputil.PostMultipart(ctx, r.client, r.url, "file", r.filename, r.contentType, content, &result); err != nil {
		return "", fmt.Errorf("extractor service: %w", err)
	}
	return result.Text, nil
}
