package extractor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestPDFExtractor_Success(t *testing.T) {
	expected := "extracted pdf text"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("could not parse multipart form: %v", err)
		}
		if r.MultipartForm.File["file"] == nil {
			t.Error("expected file field in form")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"text": expected})
	}))
	defer server.Close()

	ext := NewPDFExtractor(server.URL)
	text, err := ext.Extract(context.Background(), strings.NewReader("fake pdf content"))

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if text != expected {
		t.Errorf("expected %q, got %q", expected, text)
	}
}

func TestPDFExtractor_ServiceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"detail": "extraction failed"}`))
	}))
	defer server.Close()

	ext := NewPDFExtractor(server.URL)
	_, err := ext.Extract(context.Background(), strings.NewReader("fake pdf"))

	if err == nil {
		t.Error("expected error on 500 response, got nil")
	}
}

func TestPDFExtractor_ServiceUnavailable(t *testing.T) {
	ext := NewPDFExtractor("http://localhost:19999")
	_, err := ext.Extract(context.Background(), strings.NewReader("fake pdf"))

	if err == nil {
		t.Error("expected error when service is unreachable, got nil")
	}
}

func TestPDFExtractor_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	ext := NewPDFExtractor(server.URL)
	_, err := ext.Extract(context.Background(), strings.NewReader("fake pdf"))

	if err == nil {
		t.Error("expected error on invalid JSON response, got nil")
	}
}

func TestPDFExtractor_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ext := NewPDFExtractor(server.URL)
	_, err := ext.Extract(ctx, strings.NewReader("fake pdf"))

	if err == nil {
		t.Error("expected error on cancelled context, got nil")
	}
}
