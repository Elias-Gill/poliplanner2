package excel

import "time"

// ParseAudit records a single parse attempt, successful or not. It is the
// append-only audit trail of the ingestion pipeline.
type ParseAudit struct {
	ID        int64
	VersionID *SheetVersionID

	SourceType SourceType
	Name       string
	URL        string
	SourceDate time.Time

	StartedAt  time.Time
	FinishedAt time.Time

	Succeeded    bool
	Error        string
	ParsedSheets int
}
