package document

import (
	"mime/multipart"
	"path/filepath"
	"strings"
)

const maxFileSize = 50 << 20

var allowedTypes = map[string]string{
	".pdf":  "application/pdf",
	".txt":  "text/plain",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
}

type fileInfo struct {
	ext         string
	contentType string
	size        int64
	name        string
}

func validateFile(header *multipart.FileHeader) (*fileInfo, error) {
	if header.Size > maxFileSize {
		return nil, ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	expectedMIME, ok := allowedTypes[ext]
	if !ok {
		return nil, ErrUnsupportedType
	}

	actualMIME := header.Header.Get("Content-Type")
	if actualMIME != expectedMIME {
		return nil, ErrUnsupportedType
	}

	return &fileInfo{
		ext:         ext,
		contentType: expectedMIME,
		size:        header.Size,
		name:        header.Filename,
	}, nil
}
