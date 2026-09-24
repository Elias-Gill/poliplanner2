package excel

import (
	"context"
	"time"

	"github.com/elias-gill/poliplanner2/internal/model/excel"
)

type ExcelRepository interface {
	// SaveVersion persists a successfully parsed source version and returns its id.
	SaveVersion(ctx context.Context, version *excel.SheetVersion) (excel.SheetVersionID, error)

	// ListVersions lists the successful versions of a source type, latest first.
	ListVersions(ctx context.Context, kind excel.SourceType) ([]*excel.SheetVersion, error)

	// IsSourceUpToDate reports whether a successful version of the same source
	// (matched by name or url) with a source date at least as new already exists.
	IsSourceUpToDate(ctx context.Context, kind excel.SourceType, name, url string, sourceDate time.Time) (bool, error)

	// SaveAudit records a parse attempt, successful or not.
	SaveAudit(ctx context.Context, audit *excel.ParseAudit) error

	// ListAudit lists the parse attempts of a source type, latest first.
	ListAudit(ctx context.Context, kind excel.SourceType, limit int) ([]*excel.ParseAudit, error)
}
