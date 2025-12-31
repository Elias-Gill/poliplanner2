package store

import (
	"context"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

type SqliteSubjectStore struct {
}

func NewSqliteSubjectStore() *SqliteSubjectStore {
	return &SqliteSubjectStore{}
}

func (s SqliteSubjectStore) Insert(
	ctx context.Context,
	exec Executor,
	careerID int64,
	sub *model.Subject,
) error {
	const query = `
	INSERT INTO subjects (
		career_id, department, subject_name, semester, section,
		teacher_title, teacher_lastname, teacher_name, teacher_email,
		monday, monday_classroom,
		tuesday, tuesday_classroom,
		wednesday, wednesday_classroom,
		thursday, thursday_classroom,
		friday, friday_classroom,
		saturday, saturday_night_dates, saturday_classroom,
		partial1_date, partial1_time, partial1_classroom,
		partial2_date, partial2_time, partial2_classroom,
		final1_date, final1_time, final1_classroom,
		final1_review_date, final1_review_time,
		final2_date, final2_time, final2_classroom,
		final2_review_date, final2_review_time,
		committee_chair, committee_member1, committee_member2
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
	          ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := exec.ExecContext(
		ctx,
		query,
		careerID, sub.Department, sub.SubjectName, sub.Semester, sub.Section,
		sub.TeacherTitle, sub.TeacherLastname, sub.TeacherName, sub.TeacherEmail,
		sub.Monday, sub.MondayRoom,
		sub.Tuesday, sub.TuesdayRoom,
		sub.Wednesday, sub.WednesdayRoom,
		sub.Thursday, sub.ThursdayRoom,
		sub.Friday, sub.FridayRoom,
		sub.Saturday, sub.SaturdayDates, sub.SaturdayRoom,
		sub.Partial1Date, sub.Partial1Time, sub.Partial1Room,
		sub.Partial2Date, sub.Partial2Time, sub.Partial2Room,
		sub.Final1Date, sub.Final1Time, sub.Final1Room, sub.Final1RevDate, sub.Final1RevTime,
		sub.Final2Date, sub.Final2Time, sub.Final2Room, sub.Final2RevDate, sub.Final2RevTime,
		sub.CommitteePresident, sub.CommitteeMember1, sub.CommitteeMember2,
	)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err == nil {
		sub.ID = id
	}

	return nil
}

func (s SqliteSubjectStore) GetByID(
	ctx context.Context,
	exec Executor,
	subjectID int64,
) (*model.Subject, error) {
	const query = `
	SELECT
		subject_id, career_id, department, subject_name, semester, section,
		teacher_title, teacher_lastname, teacher_name, teacher_email,
		monday, monday_classroom,
		tuesday, tuesday_classroom,
		wednesday, wednesday_classroom,
		thursday, thursday_classroom,
		friday, friday_classroom,
		saturday, saturday_night_dates, saturday_classroom,
		partial1_date, partial1_time, partial1_classroom,
		partial2_date, partial2_time, partial2_classroom,
		final1_date, final1_time, final1_classroom,
		final1_review_date, final1_review_time,
		final2_date, final2_time, final2_classroom,
		final2_review_date, final2_review_time,
		committee_chair, committee_member1, committee_member2
	FROM subjects
	WHERE subject_id = ?
	`

	row := exec.QueryRowContext(ctx, query, subjectID)

	sub := &model.Subject{}
	err := row.Scan(
		&sub.ID, &sub.CareerID, &sub.Department, &sub.SubjectName, &sub.Semester, &sub.Section,
		&sub.TeacherTitle, &sub.TeacherLastname, &sub.TeacherName, &sub.TeacherEmail,
		&sub.Monday, &sub.MondayRoom,
		&sub.Tuesday, &sub.TuesdayRoom,
		&sub.Wednesday, &sub.WednesdayRoom,
		&sub.Thursday, &sub.ThursdayRoom,
		&sub.Friday, &sub.FridayRoom,
		&sub.Saturday, &sub.SaturdayDates, &sub.SaturdayRoom,
		&sub.Partial1Date, &sub.Partial1Time, &sub.Partial1Room,
		&sub.Partial2Date, &sub.Partial2Time, &sub.Partial2Room,
		&sub.Final1Date, &sub.Final1Time, &sub.Final1Room, &sub.Final1RevDate, &sub.Final1RevTime,
		&sub.Final2Date, &sub.Final2Time, &sub.Final2Room, &sub.Final2RevDate, &sub.Final2RevTime,
		&sub.CommitteePresident, &sub.CommitteeMember1, &sub.CommitteeMember2,
	)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (s SqliteSubjectStore) GetByCareerID(
	ctx context.Context,
	exec Executor,
	careerID int64,
) ([]*SubjectListItem, error) {
	const query = `
	SELECT
		subject_id,
		subject_name,
		semester,
		section,
		teacher_title,
		teacher_name,
		teacher_lastname
	FROM subjects
	WHERE career_id = ?
	ORDER BY semester, subject_name
	`

	rows, err := exec.QueryContext(ctx, query, careerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*SubjectListItem, 0, 64)

	for rows.Next() {
		var item SubjectListItem

		err := rows.Scan(
			&item.ID, &item.SubjectName, &item.Semester, &item.Section,
			&item.TeacherTitle, &item.TeacherName, &item.TeacherLastname,
		)
		if err != nil {
			return nil, err
		}

		result = append(result, &item)
	}

	return result, rows.Err()
}

func (s SqliteSubjectStore) FindEquivalentSubjectIDBySheetVersion(
	ctx context.Context,
	exec Executor,
	subjectName string,
	section string,
	sheetVersionID int64,
) (int64, error) {

	query := `
	SELECT s.subject_id
	FROM subjects s
	JOIN career_version cv ON cv.career_version_id = s.career_id
	JOIN career c ON c.career_id = cv.career_id
	WHERE s.subject_name = ?
	  AND cv.sheet_version_id = ?
	  AND s.section = ?
	LIMIT 1
	`

	var subjectID int64 = 0
	err := exec.QueryRowContext(
		ctx,
		query,
		subjectName,
		sheetVersionID,
		section,
	).Scan(&subjectID)

	return subjectID, err
}
