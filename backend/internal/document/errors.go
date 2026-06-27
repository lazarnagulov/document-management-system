package document

import "errors"

var (
	ErrNotFound        = errors.New("document not found")
	ErrForbidden       = errors.New("access denied")
	ErrInvalidFile     = errors.New("invalid file")
	ErrUnsupportedType = errors.New("unsupported file type")
	ErrFileTooLarge    = errors.New("file too large")
)
