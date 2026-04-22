package main

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	excelimport "github.com/elias-gill/poliplanner2/internal/app/excelImport"
	scheduleApp "github.com/elias-gill/poliplanner2/internal/app/schedule"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/schedule"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/persistence"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/scraper"
	"github.com/elias-gill/poliplanner2/logger"
)

// WARNING: This assumes the server has ALL new db migrations up to date before this is runned
// WARNING: This script performs irreversible data writes. Test thoroughly before running in production.

func main() {
	config.MustLoad()
	logger.Warn("Starting schedules data migration")

	conn, err := persistence.ConnectDB()
	if err != nil {
		logger.Error("Cannot instantiate db connection", "error", err)
		return
	}

	db := conn.GetConnection()

	app := scheduleApp.New(sqlite.NewSqliteScheduleStore(db))

	// =========================================
	// Step 0: force new excel import
	// =========================================

	importer := excelimport.New(sqlite.NewSqliteExcelImportStore(db), sqlite.NewSqliteSheetVersionStore(db))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	if syncNewestVersion(ctx, importer) != nil {
		logger.Error("Cannot download new excel from internet")
		return
	}

	// =========================================
	// Step 1: iterate users
	// =========================================

	usersRows, err := db.Query(`
		SELECT DISTINCT user_id
		FROM schedules
	`)
	if err != nil {
		logger.Error("Cannot list users", "error", err)
		return
	}
	defer usersRows.Close()

	for usersRows.Next() {
		var userID int64

		if err := usersRows.Scan(&userID); err != nil {
			logger.Error("Cannot scan user", "error", err)
			return
		}

		// =========================================
		// Step 2: get distinct schedule names per user
		// =========================================

		namesRows, err := db.Query(`
			SELECT DISTINCT schedule_description
			FROM schedules
			WHERE user_id = ?
		`, userID)
		if err != nil {
			logger.Error("Cannot list schedule names", "error", err, "user_id", userID)
			return
		}

		var names []string

		for namesRows.Next() {
			var name string
			if err := namesRows.Scan(&name); err != nil {
				namesRows.Close()
				logger.Error("Cannot scan schedule name", "error", err, "user_id", userID)
				return
			}
			names = append(names, name)
		}
		namesRows.Close()

		// =========================================
		// Step 3: for each name, get latest schedule
		// =========================================

		for _, name := range names {

			var (
				scheduleID  int64
				description string
				createdAt   string
			)

			err := db.QueryRow(`
				SELECT schedule_id, schedule_description, created_at
				FROM schedules
				WHERE user_id = ?
				AND schedule_description = ?
				ORDER BY schedule_id DESC
				LIMIT 1
			`, userID, name).Scan(
				&scheduleID,
				&description,
				&createdAt,
			)

			if err != nil {
				logger.Error("Cannot fetch latest schedule",
					"error", err,
					"user_id", userID,
					"name", name,
				)
				continue
			}

			// =========================================
			// Step 4: retrieve old subjects
			// =========================================

			subRows, err := db.Query(`
				SELECT s.subject_name, s.section
				FROM schedule_subjects dt
				JOIN subjects s ON s.subject_id = dt.subject_id
				WHERE dt.schedule_id = ?
			`, scheduleID)
			if err != nil {
				logger.Error("Cannot query schedule subjects",
					"error", err,
					"schedule_id", scheduleID,
				)
				continue
			}

			var courses []auxCourse

			for subRows.Next() {
				var c auxCourse

				if err := subRows.Scan(&c.Name, &c.Section); err != nil {
					subRows.Close()
					logger.Error("Cannot scan subject row",
						"error", err,
						"schedule_id", scheduleID,
					)
					return
				}

				c.Name = strings.TrimSpace(c.Name)
				c.Section = strings.TrimSpace(c.Section)

				courses = append(courses, c)
			}
			subRows.Close()

			// =========================================
			// Step 5: map to new courses
			// =========================================

			var courseIDs []courseOffering.CourseOfferingID

			for _, c := range courses {
				var id int64

				err := db.QueryRow(`
					SELECT id
					FROM cursos
					WHERE nombre = ?
					AND seccion = ?
					LIMIT 1
				`, c.Name, c.Section).Scan(&id)

				if err != nil {
					if err == sql.ErrNoRows {
						continue
					}

					logger.Error("Cannot find course",
						"error", err,
						"name", c.Name,
						"section", c.Section,
						"user_id", userID,
					)
					continue
				}

				// Skip over repeated course ids
				if slices.Contains(courseIDs, courseOffering.CourseOfferingID(id)) {
					continue
				}

				courseIDs = append(courseIDs, courseOffering.CourseOfferingID(id))
			}

			// =========================================
			// Step 6: create new schedule
			// =========================================

			newSchedule, err := schedule.NewSchedule(
				user.UserID(userID),
				description,
				courseIDs,
			)
			if err != nil {
				logger.Warn("Empty schedule",
					"schedule_id", scheduleID,
					"user_id", userID,
				)
				continue
			}

			_, err = app.Save(context.Background(), *newSchedule)
			if err != nil {
				logger.Error("Cannot save schedule",
					"error", err,
					"schedule_id", scheduleID,
					"user_id", userID,
				)
				return
			}
		}
	}

	if err := usersRows.Err(); err != nil {
		logger.Error("Error iterating users", "error", err)
		return
	}

	logger.Info("Data migration completed successfully")
}

type auxCourse struct {
	Name    string
	Section string
}

func syncNewestVersion(ctx context.Context, importer *excelimport.ExcelImporter) error {
	logger.Info("Forcing Excel import (no version checks)")

	key := config.Get().Excel.GoogleAPIKey
	scraper := scraper.NewWebScraper(scraper.NewGoogleDriveHelper(key))

	source, err := scraper.FindLatestDownloadSource(ctx)
	if err != nil {
		logger.Error("Failed to find Excel source", "error", err)
		return fmt.Errorf("error fetching Excel source: %w", err)
	}

	logger.Info("Source found, proceeding with import",
		"date", source.GetMetadata().Date,
	)

	if err := importer.PersistSource(ctx, source); err != nil {
		logger.Error("Failed to persist Excel source", "error", err)
		return fmt.Errorf("error persisting Excel source: %w", err)
	}

	logger.Info("Excel import completed successfully")
	return nil
}
