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

	// SetLastSearchAttempt records a web search (discovery) of a source type. It
	// is recorded even when no new source was found.
	SetLastSearchAttempt(ctx context.Context, kind excel.SourceType, t time.Time) error

	// SetLastSyncAttempt records a persistence run of new sources for a source
	// type. It is not recorded when everything was already up to date.
	SetLastSyncAttempt(ctx context.Context, kind excel.SourceType, t time.Time) error
}
