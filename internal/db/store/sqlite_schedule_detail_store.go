package store

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqliteScheduleDetailStore struct {
	db *sql.DB
}

func NewSqliteScheduleDetailStore(db *sql.DB) *SqliteScheduleDetailStore {
	return &SqliteScheduleDetailStore{db: db}
}

func (s *SqliteScheduleDetailStore) Insert(ctx context.Context, scheduleID, subjectID int64) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO schedule_subjects (schedule_id, subject_id) VALUES (?, ?)`,
		scheduleID, subjectID,
	)
	return err
}

func (s *SqliteScheduleDetailStore) GetSubjectsByScheduleID(ctx context.Context, scheduleID int64) ([]*model.Subject, error) {
	query := `
	SELECT s.subject_id, s.career_id, s.department, s.subject_name, s.semester, s.section,
	s.teacher_title, s.teacher_lastname, s.teacher_name, s.teacher_email,
	s.monday, s.monday_classroom, s.tuesday, s.tuesday_classroom,
	s.wednesday, s.wednesday_classroom, s.thursday, s.thursday_classroom,
	s.friday, s.friday_classroom, s.saturday, s.saturday_night_dates,
	s.partial1_date, s.partial1_time, s.partial1_classroom,
	s.partial2_date, s.partial2_time, s.partial2_classroom,
	s.final1_date, s.final1_time, s.final1_classroom,
	s.final1_review_date, s.final1_review_time,
	s.final2_date, s.final2_time, s.final2_classroom,
	s.final2_review_date, s.final2_review_time,
	s.committee_chair, s.committee_member1, s.committee_member2
	FROM subjects s
	JOIN schedule_subjects ss ON s.subject_id = ss.subject_id
	WHERE ss.schedule_id = ?
	ORDER BY s.semester, s.subject_name
	`

	rows, err := s.db.QueryContext(ctx, query, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subjects []*model.Subject
	for rows.Next() {
		sub := &model.Subject{}
		var careerID sql.NullInt64
		var p1d, p2d, f1d, f1rd, f2d, f2rd sql.NullTime

		err := rows.Scan(
			&sub.SubjectID, &careerID, &sub.Department, &sub.SubjectName, &sub.Semester, &sub.Section,
			&sub.TeacherTitle, &sub.TeacherLastname, &sub.TeacherName, &sub.TeacherEmail,
			&sub.Monday, &sub.MondayRoom,
			&sub.Tuesday, &sub.TuesdayRoom,
			&sub.Wednesday, &sub.WednesdayRoom,
			&sub.Thursday, &sub.ThursdayRoom,
			&sub.Friday, &sub.FridayRoom,
			&sub.Saturday, &sub.SaturdayDates,
			&p1d, &sub.Partial1Time, &sub.Partial1Room,
			&p2d, &sub.Partial2Time, &sub.Partial2Room,
			&f1d, &sub.Final1Time, &sub.Final1Room,
			&f1rd, &sub.Final1RevTime,
			&f2d, &sub.Final2Time, &sub.Final2Room,
			&f2rd, &sub.Final2RevTime,
			&sub.CommitteePresident, &sub.CommitteeMember1, &sub.CommitteeMember2,
		)
		if err != nil {
			return nil, err
		}

		if careerID.Valid {
			sub.CareerID = careerID.Int64
		}
		if p1d.Valid {
			sub.Partial1Date = &p1d.Time
		}
		if p2d.Valid {
			sub.Partial2Date = &p2d.Time
		}
		if f1d.Valid {
			sub.Final1Date = &f1d.Time
		}
		if f1rd.Valid {
			sub.Final1RevDate = &f1rd.Time
		}
		if f2d.Valid {
			sub.Final2Date = &f2d.Time
		}
		if f2rd.Valid {
			sub.Final2RevDate = &f2rd.Time
		}

		subjects = append(subjects, sub)
	}
	return subjects, rows.Err()
}
