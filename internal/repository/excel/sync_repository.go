package excel

import (
	"context"
	"time"
)

type SyncRepository interface {
	GetLastSyncAttempt(ctx context.Context) (*time.Time, error)
	SetLastSyncAttempt(ctx context.Context, date time.Time) error
}
