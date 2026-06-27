package extractor

import "fmt"

type Factory struct {
	pdf  *PDFExtractor
	docx *DOCXExtractor
	txt  *TXTExtractor
}

func NewFactory(extractorURL string) *Factory {
	return &Factory{
		pdf:  NewPDFExtractor(extractorURL),
		docx: NewDOCXExtractor(),
		txt:  NewTXTExtractor(),
	}
}

func (f *Factory) For(contentType string) (Extractor, error) {
	switch contentType {
	case "application/pdf":
		return f.pdf, nil
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return f.docx, nil
	case "text/plain":
		return f.txt, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedType, contentType)
	}
}
