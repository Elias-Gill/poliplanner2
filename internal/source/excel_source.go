package source

import (
	"context"
	"io"
	"time"
)

type ExcelSource interface {
	GetContent(ctx context.Context) (io.ReadCloser, error)
	GetMetadata() ExcelSourceMetadata
}

type ExcelSourceMetadata struct {
	Name   string
	URI    string
	Period int
	Date   time.Time
}
