package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	txManager "github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite/tx_manager"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/internal/model/excel"
)

const dbTimeLayout = "2006-01-02 15:04:05"

type SQLiteExcelRepository struct {
	db *sql.DB
}

func NewExcelRepository(db *sql.DB) *SQLiteExcelRepository {
	return &SQLiteExcelRepository{db: db}
}

func (r *SQLiteExcelRepository) SaveVersion(ctx context.Context, version *excel.SheetVersion) (excel.SheetVersionID, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	res, err := exec.ExecContext(ctx, `
		INSERT INTO sheet_version (
			source_type,
			file_name,
			url,
			source_date,
			parsed_at,
			parsed_sheets,
			period
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
		string(version.SourceType),
		version.Name,
		version.URL,
		formatNullableTime(version.SourceDate),
		version.ParsedAt.Format(dbTimeLayout),
		version.ParsedSheets,
		nullablePeriod(version.PeriodID),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert sheet version: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to read inserted sheet version id: %w", err)
	}

	return excel.SheetVersionID(id), nil
}

func (r *SQLiteExcelRepository) ListVersions(ctx context.Context, kind excel.SourceType) ([]*excel.SheetVersion, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	rows, err := exec.QueryContext(ctx, `
		SELECT
			version_id,
			source_type,
			file_name,
			url,
			source_date,
			parsed_at,
			parsed_sheets,
			period
		FROM sheet_version
		WHERE source_type = ?
		ORDER BY source_date DESC, version_id DESC
		`, string(kind))
	if err != nil {
		return nil, fmt.Errorf("failed to query sheet versions: %w", err)
	}
	defer rows.Close()

	var versions []*excel.SheetVersion

	for rows.Next() {
		v := &excel.SheetVersion{}
		var sourceType string
		var sourceDate, parsedAt sql.NullString
		var periodID sql.NullInt64

		err := rows.Scan(
			&v.ID,
			&sourceType,
			&v.Name,
			&v.URL,
			&sourceDate,
			&parsedAt,
			&v.ParsedSheets,
			&periodID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sheet version row: %w", err)
		}

		v.SourceType = excel.SourceType(sourceType)
		v.SourceDate = parseDBTime(sourceDate)
		v.ParsedAt = parseDBTime(parsedAt)

		if periodID.Valid {
			v.PeriodID = academic.PeriodID(periodID.Int64)
		}

		versions = append(versions, v)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during sheet versions iteration: %w", err)
	}

	return versions, nil
}

func (r *SQLiteExcelRepository) IsSourceUpToDate(
	ctx context.Context,
	kind excel.SourceType,
	name, url string,
	sourceDate time.Time,
) (bool, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	var exists int
	err := exec.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM sheet_version
			WHERE source_type = ?
			  AND (file_name = ? OR url = ?)
			  AND source_date >= ?
		)
		`,
		string(kind),
		name,
		url,
		formatNullableTime(sourceDate),
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check source version: %w", err)
	}

	return exists == 1, nil
}

func (r *SQLiteExcelRepository) SaveAudit(ctx context.Context, audit *excel.ParseAudit) error {
	exec := txManager.GetExecutor(ctx, r.db)

	var versionID any
	if audit.VersionID != nil {
		versionID = int64(*audit.VersionID)
	}

	_, err := exec.ExecContext(ctx, `
		INSERT INTO sheet_parse_audit (
			version_id,
			source_type,
			file_name,
			url,
			source_date,
			started_at,
			finished_at,
			success,
			error_message,
			parsed_sheets
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
		versionID,
		string(audit.SourceType),
		audit.Name,
		audit.URL,
		formatNullableTime(audit.SourceDate),
		audit.StartedAt.Format(dbTimeLayout),
		audit.FinishedAt.Format(dbTimeLayout),
		boolToInt(audit.Succeeded),
		audit.Error,
		audit.ParsedSheets,
	)
	if err != nil {
		return fmt.Errorf("failed to insert parse audit: %w", err)
	}

	return nil
}

func (r *SQLiteExcelRepository) ListAudit(ctx context.Context, kind excel.SourceType, limit int) ([]*excel.ParseAudit, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	if limit <= 0 {
		limit = 100
	}

	rows, err := exec.QueryContext(ctx, `
		SELECT
			audit_id,
			version_id,
			source_type,
			file_name,
			url,
			source_date,
			started_at,
			finished_at,
			success,
			error_message,
			parsed_sheets
		FROM sheet_parse_audit
		WHERE source_type = ?
		ORDER BY audit_id DESC
		LIMIT ?
		`, string(kind), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query parse audit: %w", err)
	}
	defer rows.Close()

	var audits []*excel.ParseAudit

	for rows.Next() {
		a := &excel.ParseAudit{}
		var versionID sql.NullInt64
		var sourceType string
		var sourceDate, startedAt, finishedAt sql.NullString
		var success int
		var errorMessage sql.NullString

		err := rows.Scan(
			&a.ID,
			&versionID,
			&sourceType,
			&a.Name,
			&a.URL,
			&sourceDate,
			&startedAt,
			&finishedAt,
			&success,
			&errorMessage,
			&a.ParsedSheets,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan parse audit row: %w", err)
		}

		if versionID.Valid {
			id := excel.SheetVersionID(versionID.Int64)
			a.VersionID = &id
		}

		a.SourceType = excel.SourceType(sourceType)
		a.SourceDate = parseDBTime(sourceDate)
		a.StartedAt = parseDBTime(startedAt)
		a.FinishedAt = parseDBTime(finishedAt)
		a.Succeeded = success == 1

		if errorMessage.Valid {
			a.Error = errorMessage.String
		}

		audits = append(audits, a)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during parse audit iteration: %w", err)
	}

	return audits, nil
}

func formatNullableTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}

	return t.Format(dbTimeLayout)
}

func parseDBTime(value sql.NullString) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	parsed, err := time.Parse(dbTimeLayout, value.String)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339, value.String)
	}

	if err != nil {
		return time.Time{}
	}

	return parsed
}

func nullablePeriod(id academic.PeriodID) any {
	if id == 0 {
		return nil
	}

	return int64(id)
}

func boolToInt(value bool) int {
	if value {
		return 1
	}

	return 0
}
