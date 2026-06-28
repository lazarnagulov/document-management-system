package extractor

import (
	"context"
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type TXTExtractor struct{}

func NewTXTExtractor() *TXTExtractor {
	return &TXTExtractor{}
}

func (e *TXTExtractor) Extract(ctx context.Context, content io.Reader) (string, error) {
	data, err := io.ReadAll(content)
	if err != nil {
		return "", err
	}
	if isValidUTF8(data) {
		return strings.TrimSpace(string(data)), nil
	}

	decoded, err := decodeLatin2(data)
	if err != nil {
		return "", fmt.Errorf("decode text file: %w", err)
	}

	return strings.TrimSpace(decoded), nil
}

func isValidUTF8(data []byte) bool {
	return string(data) == strings.ToValidUTF8(string(data), "")
}

func decodeLatin2(data []byte) (string, error) {
	decoder := charmap.ISO8859_2.NewDecoder()
	result, _, err := transform.Bytes(decoder, data)
	if err != nil {
		return "", err
	}
	return string(result), nil
}
