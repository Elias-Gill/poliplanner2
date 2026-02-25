package source

import (
	"context"
	"io"
)

// ReaderExcelSource is a basic implementation of the ExcelSource interface.
// It is intended for manually provided Excel files (e.g., uploaded via a web form)
// and wraps an io.ReadCloser for the file content along with metadata like name, URI, period, and upload date.
type ReaderExcelSource struct {
	Reader   io.ReadCloser
	Metadata ExcelSourceMetadata
}

func NewExcelSourceFromReader(reader io.ReadCloser, meta ExcelSourceMetadata) ExcelSource {
	return &ReaderExcelSource{
		Reader:   reader,
		Metadata: meta,
	}
}

func (m *ReaderExcelSource) GetContent(ctx context.Context) (io.ReadCloser, error) {
	return m.Reader, nil
}

func (m *ReaderExcelSource) GetMetadata() ExcelSourceMetadata {
	return m.Metadata
}
