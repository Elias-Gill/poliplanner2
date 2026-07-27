package excel

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/model/excel"
)

type ExcelRepository interface {
	SaveVersion(ctx context.Context, version *excel.SheetVersion) error

	// Lists all parsed excel versions ordered by date (latest to oldest)
	ListAllVersions(ctx context.Context) ([]*excel.SheetVersion, error)
}
