package excel

import "time"

// SyncState tracks the search and sync activity per source type, so the server
// can avoid unnecessary scrapes.
type SyncState struct {
	SourceType SourceType

	// LastSearchAt is when the server last scraped the web for this source type.
	// Recorded after a successful discovery, even when nothing new was found.
	// AutoSync uses it to decide whether to search again.
	LastSearchAt *time.Time

	// LastSyncAt is when the server last parsed and persisted new sources of
	// this type. Recorded only when there was something pending to sync.
	LastSyncAt *time.Time
}
