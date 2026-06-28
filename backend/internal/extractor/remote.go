package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type remoteExtractor struct {
	url      string
	filename string
	client   *http.Client
}

func newRemoteExtractor(url, filename string) remoteExtractor {
	return remoteExtractor{
		url:      url,
		filename: filename,
		client:   &http.Client{},
	}
}

func (r *remoteExtractor) extract(ctx context.Context, content io.Reader) (string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", r.filename)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, content); err != nil {
		return "", fmt.Errorf("copy content: %w", err)
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url, body)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := r.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call extractor service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("extractor service error %d: %s", resp.StatusCode, raw)
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Text, nil
}
