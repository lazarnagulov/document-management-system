package extractor

import "fmt"

type Factory struct {
	pdf  *PDFExtractor
	docx *DOCXExtractor
	txt  *TXTExtractor
}

func NewFactory(cfg Config) *Factory {
	return &Factory{
		pdf:  NewPDFExtractor(cfg.PDFURL),
		docx: NewDOCXExtractor(cfg.DOCXURL),
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
