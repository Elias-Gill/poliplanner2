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
	// Date represents the source date of the data.
	// - In the case of scraped Excel files, it corresponds to the date indicated in
	// 	the file name.
	// - For manually uploaded files via an endpoint, it corresponds to the date when
	// 	the file was uploaded.
	Date time.Time
}
