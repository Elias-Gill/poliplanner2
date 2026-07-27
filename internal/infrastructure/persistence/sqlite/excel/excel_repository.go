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

type SQLiteExcelRepository struct {
	db *sql.DB
}

func NewExcelRepository(db *sql.DB) *SQLiteExcelRepository {
	return &SQLiteExcelRepository{db: db}
}

func (r *SQLiteExcelRepository) SaveVersion(ctx context.Context, version *excel.SheetVersion) error {
	exec := txManager.GetExecutor(ctx, r.db)

	successInt := 1
	if version.Error != "" {
		successInt = 0
	}

	_, err := exec.ExecContext(ctx, `
		INSERT INTO sheet_version (
			file_name,
			url,
			success,
			error_message,
			parsed_sheets,
			period,
			parsed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
		version.Name,
		version.URL,
		successInt,
		version.Error,
		version.ParsedSheets,
		version.PeriodID,
		version.ParsedAt.Format("2006-01-02 15:04:05"),
	)

	if err != nil {
		return fmt.Errorf("failed to insert sheet version: %w", err)
	}

	return nil
}

func (r *SQLiteExcelRepository) ListAllVersions(ctx context.Context) ([]*excel.SheetVersion, error) {
	exec := txManager.GetExecutor(ctx, r.db)

	rows, err := exec.QueryContext(ctx, `
        SELECT 
            version_id, 
            file_name, 
            url, 
            success, 
            error_message, 
            parsed_sheets, 
            period, 
            parsed_at
        FROM sheet_version
        ORDER BY parsed_at DESC
        `)
	if err != nil {
		return nil, fmt.Errorf("failed to query sheet versions: %w", err)
	}
	defer rows.Close()

	var versions []*excel.SheetVersion

	for rows.Next() {
		v := &excel.SheetVersion{}
		var successInt int
		var parsedAtStr string
		var errorMessage sql.NullString
		var periodID sql.NullInt64

		err := rows.Scan(
			&v.ID,
			&v.Name,
			&v.URL,
			&successInt,
			&errorMessage,
			&v.ParsedSheets,
			&periodID,
			&parsedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sheet version row: %w", err)
		}

		v.Succeeded = successInt == 1

		if errorMessage.Valid {
			v.Error = errorMessage.String
		}

		if periodID.Valid {
			v.PeriodID = academic.PeriodID(periodID.Int64)
		}

		parsedAt, err := time.Parse("2006-01-02 15:04:05", parsedAtStr)
		if err != nil {
			parsedAt, err = time.Parse(time.RFC3339, parsedAtStr)
		}

		if err == nil {
			v.ParsedAt = parsedAt
		}

		versions = append(versions, v)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during sheet versions iteration: %w", err)
	}

	return versions, nil
}
