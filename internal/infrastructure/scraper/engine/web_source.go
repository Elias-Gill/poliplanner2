package engine

import "github.com/elias-gill/poliplanner2/internal/infrastructure/source"

// WebSource is a source backed by a remote Excel file. It is an alias of the
// shared source.URLSource, so the scraper and the manual ingestion path use the
// exact same download and metadata logic.
type WebSource = source.URLSource
