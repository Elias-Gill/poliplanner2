package excel

import "time"

// SyncState tracks the search and sync activity per source type, so the server
// can avoid unnecessary scrapes.
type SyncState struct {
	SourceType SourceType

	LastSearchAt *time.Time
	LastSyncAt   *time.Time
}
