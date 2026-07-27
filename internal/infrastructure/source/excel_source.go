package source

import (
	"context"
	"io"
	"time"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type SourceScraper interface {
	Discover(ctx context.Context) ([]ExcelSource, error)
}

type ExcelSource interface {
	Content(ctx context.Context) (io.ReadCloser, error)
	Metadata() ExcelSourceMetadata
}

type ExcelSourceMetadata struct {
	Name string

	URI string

	// Semester represents the current academic period of the year.
	// The university operates on a semestral cycle calculated as follows:
	//
	//	- First Semester:  August to December
	//	- Second Semester: January to July
	Semester academic.YearSemester

	// Date identifies the version of the source data.
	//
	// For scraped Excel files, it is the date encoded in the file name.
	// For manually uploaded files, it is the upload date.
	//
	// This value is used to determine whether the source is newer than the
	// latest successfully parsed version stored in the database.
	Date time.Time
}
