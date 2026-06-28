package extractor

import (
	"context"
	"io"
)

type DOCXExtractor struct {
	remoteExtractor
}

func NewDOCXExtractor(url string) *DOCXExtractor {
	return &DOCXExtractor{newRemoteExtractor(url+"/extract", "document.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")}
}

func (e *DOCXExtractor) Extract(ctx context.Context, content io.Reader) (string, error) {
	return e.extract(ctx, content)
}
