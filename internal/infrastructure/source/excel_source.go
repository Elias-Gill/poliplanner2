package source

import (
	"context"
	"io"
	"time"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

type SourceMetadata struct {
	Name string

	URI string

	// Semester represents the current academic period of the year.
	// The university operates on a semestral cycle calculated as follows:
	//
	//	- First Semester:  January to July
	//	- Second Semester: August to December
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

type sourceInterface interface {
	Content(ctx context.Context) (io.ReadCloser, error)
	Metadata() SourceMetadata
}

type ScheduleSource sourceInterface

type LabSource sourceInterface

type SourceScraper interface {
	DiscoverSchedules(ctx context.Context) ([]ScheduleSource, error)
	DiscoverLabs(ctx context.Context) ([]LabSource, error)
}
