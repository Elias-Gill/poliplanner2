package store

import (
	"context"
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqliteSubjectStore struct {
}

func NewSqliteSubjectStore() *SqliteSubjectStore {
	return &SqliteSubjectStore{}
}

func (s SqliteSubjectStore) Insert(ctx context.Context, exec Executor, sub *model.Subject) error {
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

	res, err := exec.ExecContext(ctx, query,
		sub.CareerID, sub.Department, sub.SubjectName, sub.Semester, sub.Section,
		sub.TeacherTitle, sub.TeacherLastname, sub.TeacherName, sub.TeacherEmail,
		sub.Monday, sub.MondayRoom,
		sub.Tuesday, sub.TuesdayRoom,
		sub.Wednesday, sub.WednesdayRoom,
		sub.Thursday, sub.ThursdayRoom,
		sub.Friday, sub.FridayRoom,
		sub.Saturday, sub.SaturdayDates,
		sub.Partial1Date, sub.Partial1Time, sub.Partial1Room,
		sub.Partial2Date, sub.Partial2Time, sub.Partial2Room,
		sub.Final1Date, sub.Final1Time, sub.Final1Room,
		sub.Final1RevDate, sub.Final1RevTime,
		sub.Final2Date, sub.Final2Time, sub.Final2Room,
		sub.Final2RevDate, sub.Final2RevTime,
		sub.CommitteePresident, sub.CommitteeMember1, sub.CommitteeMember2,
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	sub.ID = id
	return nil
}

func (s SqliteSubjectStore) GetByID(ctx context.Context, exec Executor, subjectID int64) (*model.Subject, error) {
	return s.scanOne(ctx, exec, `WHERE subject_id = ?`, subjectID)
}

func (s SqliteSubjectStore) GetByCareerID(ctx context.Context, exec Executor, careerID int64) ([]*model.Subject, error) {
	return s.scanMany(ctx, exec, `WHERE career_id = ? ORDER BY semester, subject_name`, careerID)
}

func (s SqliteSubjectStore) scanOne(ctx context.Context, exec Executor, where string, args ...any) (*model.Subject, error) {
	subs, err := s.scanMany(ctx, exec, where, args...)
	if err != nil || len(subs) == 0 {
		return nil, sql.ErrNoRows
	}
	return subs[0], nil
}

func (s SqliteSubjectStore) scanMany(ctx context.Context, exec Executor, where string, args ...any) ([]*model.Subject, error) {
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

	rows, err := exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.Subject
	for rows.Next() {
		sub := &model.Subject{}
		var careerID sql.NullInt64

		err := rows.Scan(
			&sub.ID, &careerID, &sub.Department, &sub.SubjectName, &sub.Semester, &sub.Section,
			&sub.TeacherTitle, &sub.TeacherLastname, &sub.TeacherName, &sub.TeacherEmail,
			&sub.Monday, &sub.MondayRoom,
			&sub.Tuesday, &sub.TuesdayRoom,
			&sub.Wednesday, &sub.WednesdayRoom,
			&sub.Thursday, &sub.ThursdayRoom,
			&sub.Friday, &sub.FridayRoom,
			&sub.Saturday, &sub.SaturdayDates,
			&sub.Partial1Date, &sub.Partial1Time, &sub.Partial1Room,
			&sub.Partial2Date, &sub.Partial2Time, &sub.Partial2Room,
			&sub.Final1Date, &sub.Final1Time, &sub.Final1Room,
			&sub.Final1RevDate, &sub.Final1RevTime,
			&sub.Final2Date, &sub.Final2Time, &sub.Final2Room,
			&sub.Final2RevDate, &sub.Final2RevTime,
			&sub.CommitteePresident, &sub.CommitteeMember1, &sub.CommitteeMember2,
		)
		if err != nil {
			return nil, err
		}

		if careerID.Valid {
			sub.CareerID = careerID.Int64
		}

		result = append(result, sub)
	}
	return result, rows.Err()
}
