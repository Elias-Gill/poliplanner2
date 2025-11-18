package store

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/db/models"
)

type SqliteSubjectStore struct {
	db *sql.DB
}

func NewSqliteSubjectStore(db *sql.DB) *SqliteSubjectStore {
	return &SqliteSubjectStore{db: db}
}

func (s *SqliteSubjectStore) Insert(ctx context.Context, sub *models.Subject) error {
	query := `
	INSERT INTO subjects (
	career_id, department, subject_name, semester, section,
	teacher_title, teacher_lastname, teacher_name, teacher_email,
	monday, monday_classroom,
	tuesday, tuesday_classroom,
	wednesday, wednesday_classroom,
	thursday, thursday_classroom,
	friday, friday_classroom,
	saturday, saturday_night_dates,
	partial1_date, partial1_time, partial1_classroom,
	partial2_date, partial2_time, partial2_classroom,
	final1_date, final1_time, final1_classroom,
	final1_review_date, final1_review_time,
	final2_date, final2_time, final2_classroom,
	final2_review_date, final2_review_time,
	committee_chair, committee_member1, committee_member2
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var p1d, p2d, f1d, f1rd, f2d, f2rd sql.NullTime
	if sub.Partial1Date != nil {
		p1d = sql.NullTime{Time: *sub.Partial1Date, Valid: true}
	}
	if sub.Partial2Date != nil {
		p2d = sql.NullTime{Time: *sub.Partial2Date, Valid: true}
	}
	if sub.Final1Date != nil {
		f1d = sql.NullTime{Time: *sub.Final1Date, Valid: true}
	}
	if sub.Final1RevDate != nil {
		f1rd = sql.NullTime{Time: *sub.Final1RevDate, Valid: true}
	}
	if sub.Final2Date != nil {
		f2d = sql.NullTime{Time: *sub.Final2Date, Valid: true}
	}
	if sub.Final2RevDate != nil {
		f2rd = sql.NullTime{Time: *sub.Final2RevDate, Valid: true}
	}

	res, err := s.db.ExecContext(ctx, query,
		sub.CareerID, sub.Department, sub.SubjectName, sub.Semester, sub.Section,
		sub.TeacherTitle, sub.TeacherLastname, sub.TeacherName, sub.TeacherEmail,
		sub.Monday, sub.MondayRoom,
		sub.Tuesday, sub.TuesdayRoom,
		sub.Wednesday, sub.WednesdayRoom,
		sub.Thursday, sub.ThursdayRoom,
		sub.Friday, sub.FridayRoom,
		sub.Saturday, sub.SaturdayDates,
		p1d, sub.Partial1Time, sub.Partial1Room,
		p2d, sub.Partial2Time, sub.Partial2Room,
		f1d, sub.Final1Time, sub.Final1Room,
		f1rd, sub.Final1RevTime,
		f2d, sub.Final2Time, sub.Final2Room,
		f2rd, sub.Final2RevTime,
		sub.CommitteePresident, sub.CommitteeMember1, sub.CommitteeMember2,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	sub.SubjectID = id
	return nil
}

func (s *SqliteSubjectStore) GetByID(ctx context.Context, subjectID int64) (*models.Subject, error) {
	return s.scanOne(ctx, `WHERE subject_id = ?`, subjectID)
}

func (s *SqliteSubjectStore) GetByCareerID(ctx context.Context, careerID int64) ([]*models.Subject, error) {
	return s.scanMany(ctx, `WHERE career_id = ? ORDER BY semester, subject_name`, careerID)
}

func (s *SqliteSubjectStore) scanOne(ctx context.Context, where string, args ...any) (*models.Subject, error) {
	subs, err := s.scanMany(ctx, where, args...)
	if err != nil || len(subs) == 0 {
		return nil, sql.ErrNoRows
	}
	return subs[0], nil
}

func (s *SqliteSubjectStore) scanMany(ctx context.Context, where string, args ...any) ([]*models.Subject, error) {
	query := `
	SELECT subject_id, career_id, department, subject_name, semester, section,
	teacher_title, teacher_lastname, teacher_name, teacher_email,
	monday, monday_classroom, tuesday, tuesday_classroom,
	wednesday, wednesday_classroom, thursday, thursday_classroom,
	friday, friday_classroom, saturday, saturday_night_dates,
	partial1_date, partial1_time, partial1_classroom,
	partial2_date, partial2_time, partial2_classroom,
	final1_date, final1_time, final1_classroom,
	final1_review_date, final1_review_time,
	final2_date, final2_time, final2_classroom,
	final2_review_date, final2_review_time,
	committee_chair, committee_member1, committee_member2
	FROM subjects ` + where

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Subject
	for rows.Next() {
		sub := &models.Subject{}
		var careerID sql.NullInt64
		var partial1Date, partial2Date, final1Date, final1RevDate, final2Date, final2RevDate sql.NullTime

		err := rows.Scan(
			&sub.SubjectID, &careerID, &sub.Department, &sub.SubjectName, &sub.Semester, &sub.Section,
			&sub.TeacherTitle, &sub.TeacherLastname, &sub.TeacherName, &sub.TeacherEmail,
			&sub.Monday, &sub.MondayRoom,
			&sub.Tuesday, &sub.TuesdayRoom,
			&sub.Wednesday, &sub.WednesdayRoom,
			&sub.Thursday, &sub.ThursdayRoom,
			&sub.Friday, &sub.FridayRoom,
			&sub.Saturday, &sub.SaturdayDates,
			&partial1Date, &sub.Partial1Time, &sub.Partial1Room,
			&partial2Date, &sub.Partial2Time, &sub.Partial2Room,
			&final1Date, &sub.Final1Time, &sub.Final1Room,
			&final1RevDate, &sub.Final1RevTime,
			&final2Date, &sub.Final2Time, &sub.Final2Room,
			&final2RevDate, &sub.Final2RevTime,
			&sub.CommitteePresident, &sub.CommitteeMember1, &sub.CommitteeMember2,
		)
		if err != nil {
			return nil, err
		}

		if careerID.Valid {
			sub.CareerID = careerID.Int64
		}
		if partial1Date.Valid {
			sub.Partial1Date = &partial1Date.Time
		}
		if partial2Date.Valid {
			sub.Partial2Date = &partial2Date.Time
		}
		if final1Date.Valid {
			sub.Final1Date = &final1Date.Time
		}
		if final1RevDate.Valid {
			sub.Final1RevDate = &final1RevDate.Time
		}
		if final2Date.Valid {
			sub.Final2Date = &final2Date.Time
		}
		if final2RevDate.Valid {
			sub.Final2RevDate = &final2RevDate.Time
		}

		result = append(result, sub)
	}
	return result, rows.Err()
}
