package excel

import (
	"context"
	"time"

	"github.com/elias-gill/poliplanner2/internal/model/excel"
)

type SyncRepository interface {
	// GetSyncState returns the state of a source type. When there is no row yet
	// it returns a state with nil timestamps instead of an error.
	GetSyncState(ctx context.Context, kind excel.SourceType) (*excel.SyncState, error)

	SetLastSearchAttempt(ctx context.Context, kind excel.SourceType, t time.Time) error

	SetLastSyncAttempt(ctx context.Context, kind excel.SourceType, t time.Time) error
}
