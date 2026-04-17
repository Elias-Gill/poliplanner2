package main

import (
	"context"
	"database/sql"
	"strings"

	scheduleApp "github.com/elias-gill/poliplanner2/internal/app/schedule"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	"github.com/elias-gill/poliplanner2/internal/domain/schedule"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/persistence"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/persistence/sqlite"
	"github.com/elias-gill/poliplanner2/logger"
)

// WARNING: This assumes the server has been run at least once before executing this script.
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

	rows, err := db.Query(`
		SELECT
			schedule_id,
			user_id,
			schedule_description,
			schedule_sheet_version,
			created_at
		FROM schedules
		ORDER BY schedule_id`)
	if err != nil {
		logger.Error("Error listing user schedules", "error", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			scheduleID   int64
			userID       int64
			description  string
			sheetVersion int64
			createdAt    string
		)

		err := rows.Scan(
			&scheduleID,
			&userID,
			&description,
			&sheetVersion,
			&createdAt,
		)
		if err != nil {
			logger.Error("Cannot scan schedules row", "error", err)
			return
		}

		// =========================================
		// Retrieve subjects from the old schedule
		// =========================================

		query := `
			SELECT s.subject_name, s.section
			FROM schedule_subjects dt
			JOIN subjects s ON s.subject_id = dt.subject_id
			WHERE dt.schedule_id = ?
		`

		rows2, err := db.Query(query, scheduleID)
		if err != nil {
			logger.Error("Cannot query schedule subjects", "error", err, "schedule_id", scheduleID)
			return
		}

		var courses = make([]auxCourse, 0, 8)

		for rows2.Next() {
			var c auxCourse

			err := rows2.Scan(&c.Name, &c.Section)
			if err != nil {
				rows2.Close()
				logger.Error("Cannot scan schedules row", "error", err, "schedule_id", scheduleID)
				return
			}

			// minimal defensive cleanup
			c.Name = strings.TrimSpace(c.Name)
			c.Section = strings.TrimSpace(c.Section)

			courses = append(courses, c)
		}

		if err := rows2.Err(); err != nil {
			rows2.Close()
			logger.Error("Error iterating schedule subjects rows", "error", err, "schedule_id", scheduleID)
			return
		}

		rows2.Close()

		// =========================================
		// Map to new courses
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

				logger.Error("Cannot query course",
					"error", err,
					"name", c.Name,
					"section", c.Section,
					"schedule_id", scheduleID,
				)
				return
			}

			courseIDs = append(courseIDs, courseOffering.CourseOfferingID(id))
		}

		// =========================================
		// Create new schedule
		// =========================================

		newSchedule, err := schedule.NewSchedule(
			user.UserID(userID),
			description,
			courseIDs,
		)
		if err != nil {
			logger.Error("Cannot create schedule domain object", "error", err, "schedule_id", scheduleID)
			return
		}

		_, err = app.Save(context.Background(), *newSchedule)
		if err != nil {
			logger.Error("Cannot save new schedule", "error", err, "schedule_id", scheduleID)
			return
		}
	}

	if err := rows.Err(); err != nil {
		logger.Error("Cannot retrieve schedules row", "error", err)
		return
	}

	logger.Info("Data migration completed successfully")
}

type auxCourse struct {
	Name    string
	Section string
}
