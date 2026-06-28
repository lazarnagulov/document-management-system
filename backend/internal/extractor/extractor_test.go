package extractor

import (
	"context"
	"strings"
	"testing"
)

func TestTXTExtractor_UTF8(t *testing.T) {
	ext := NewTXTExtractor()
	input := "Hello, world!\nThis is a test."

	text, err := ext.Extract(context.Background(), strings.NewReader(input))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if text != input {
		t.Errorf("expected %q, got %q", input, text)
	}
}

func TestTXTExtractor_TrimsWhitespace(t *testing.T) {
	ext := NewTXTExtractor()

	text, err := ext.Extract(context.Background(), strings.NewReader("  hello  \n"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if text != "hello" {
		t.Errorf("expected \"hello\", got %q", text)
	}
}

func TestTXTExtractor_Latin2(t *testing.T) {
	ext := NewTXTExtractor()

	latin2Input := []byte{0xe8, 0xb9, 0xbe}
	text, err := ext.Extract(context.Background(), strings.NewReader(string(latin2Input)))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if text == "" {
		t.Error("expected non-empty text from Latin-2 input")
	}
}

func TestTXTExtractor_EmptyFile(t *testing.T) {
	ext := NewTXTExtractor()

	text, err := ext.Extract(context.Background(), strings.NewReader(""))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if text != "" {
		t.Errorf("expected empty string, got %q", text)
	}
}
