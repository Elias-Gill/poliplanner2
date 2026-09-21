package source

import (
	"context"
	"io"
)

// ReaderSource is a basic Source implementation backed by an already opened
// reader. It is intended for manually provided Excel files (e.g., uploaded via a
// web form) and wraps an io.ReadCloser for the file content along with metadata
// like name, URI, period, and upload date.
type ReaderSource struct {
	reader   io.ReadCloser
	metadata SourceMetadata
}

func (m *ReaderSource) Content(ctx context.Context) (io.ReadCloser, error) {
	return m.reader, nil
}

func (m *ReaderSource) Metadata() SourceMetadata {
	return m.metadata
}

// NewScheduleSourceFromReader builds a schedule source from an uploaded file.
func NewScheduleSourceFromReader(reader io.ReadCloser, meta SourceMetadata) ScheduleSource {
	return &ReaderSource{
		reader:   reader,
		metadata: meta,
	}
}

// NewLabSourceFromReader builds a laboratory source from an uploaded file.
func NewLabSourceFromReader(reader io.ReadCloser, meta SourceMetadata) LabSource {
	return &ReaderSource{
		reader:   reader,
		metadata: meta,
	}
}
