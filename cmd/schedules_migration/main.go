package main

import (
	"context"
	"slices"
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

// WARNING: This assumes the server has ALL new db migrations up to date before this is run
// WARNING: This script performs irreversible data writes. Test thoroughly before running in production
// WARNING: cursos table must already be populated from latest Excel import

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
		// Step 2: get distinct schedule names
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
				logger.Error("Cannot scan schedule name", "error", err)
				return
			}
			names = append(names, name)
		}
		namesRows.Close()

		// =========================================
		// Step 3: process each schedule name
		// =========================================

		for _, name := range names {

			var scheduleID int64

			// Get latest schedule for this user + name
			err := db.QueryRow(`
				SELECT schedule_id
				FROM schedules
				WHERE user_id = ?
				AND schedule_description = ?
				ORDER BY schedule_id DESC
				LIMIT 1
			`, userID, name).Scan(&scheduleID)

			if err != nil {
				logger.Error("Cannot fetch latest schedule",
					"error", err,
					"user_id", userID,
					"name", name,
				)
				continue
			}

			// =========================================
			// Step 4: retrieve old subjects + career
			// =========================================

			subRows, err := db.Query(`
				SELECT 
					s.subject_name,
					s.section,
					c.career_code
				FROM schedule_subjects dt
				JOIN subjects s ON s.subject_id = dt.subject_id
				JOIN career_version cv ON cv.career_version_id = s.career_id
				JOIN career c ON c.career_id = cv.career_id
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

				if err := subRows.Scan(&c.Name, &c.Section, &c.Career); err != nil {
					subRows.Close()
					logger.Error("Cannot scan subject row",
						"error", err,
						"schedule_id", scheduleID,
					)
					return
				}

				// Minimal normalization
				c.Name = strings.TrimSpace(c.Name)
				c.Section = strings.TrimSpace(c.Section)
				c.Career = strings.ToUpper(strings.TrimSpace(c.Career))

				courses = append(courses, c)
			}

			if err := subRows.Err(); err != nil {
				subRows.Close()
				logger.Error("Error iterating subject rows", "error", err)
				return
			}

			subRows.Close()

			// =========================================
			// Step 5: map to new cursos (by career)
			// =========================================

			var courseIDs []courseOffering.CourseOfferingID

			for _, c := range courses {
				var id int64

				err := db.QueryRow(`
					SELECT cu.id
					FROM cursos cu
					JOIN mallas m ON m.id = cu.malla
					JOIN carreras ca ON ca.id = m.carrera
					WHERE cu.nombre = ?
					AND cu.seccion = ?
					AND ca.siglas = ?
					LIMIT 1
				`, c.Name, c.Section, c.Career).Scan(&id)

				if err != nil {
					logger.Warn("Error finding course",
						"error", err,
						"name", c.Name,
						"section", c.Section,
						"career", c.Career,
						"user_id", userID,
					)
					continue
				}

				// Deduplicate
				if slices.Contains(courseIDs, courseOffering.CourseOfferingID(id)) {
					continue
				}

				courseIDs = append(courseIDs, courseOffering.CourseOfferingID(id))
			}

			// =========================================
			// Step 6: create new schedule via domain
			// =========================================

			newSchedule, err := schedule.NewSchedule(
				user.UserID(userID),
				name,
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
	Career  string
}
