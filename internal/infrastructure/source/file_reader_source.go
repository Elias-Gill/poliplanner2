package source

import (
	"context"
	"io"
)

// ReaderExcelSource is a basic implementation of the ExcelSource interface.
// It is intended for manually provided Excel files (e.g., uploaded via a web form)
// and wraps an io.ReadCloser for the file content along with metadata like name, URI, period, and upload date.
type ReaderExcelSource struct {
	reader   io.ReadCloser
	metadata SourceMetadata
}

func NewExcelSourceFromReader(reader io.ReadCloser, meta SourceMetadata) ScheduleSource {
	return &ReaderExcelSource{
		reader:   reader,
		metadata: meta,
	}
}

func (m *ReaderExcelSource) Content(ctx context.Context) (io.ReadCloser, error) {
	return m.reader, nil
}

func (m *ReaderExcelSource) Metadata() SourceMetadata {
	return m.metadata
}
