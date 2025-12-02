package model

import (
	"database/sql"
)

// Subject represents the final domain model (equivalent to Java's Subject)
type Subject struct {
	SubjectID int64
	CareerID  int64

	// General info
	Department  string
	SubjectName string
	Semester    int
	Section     string

	// Teacher info
	TeacherTitle    string
	TeacherLastname string
	TeacherName     string
	TeacherEmail    string

	// Weekly schedule
	Monday    sql.NullString
	Tuesday   sql.NullString
	Wednesday sql.NullString
	Thursday  sql.NullString
	Friday    sql.NullString
	Saturday  sql.NullString

	// Classrooms
	MondayRoom    sql.NullString
	TuesdayRoom   sql.NullString
	WednesdayRoom sql.NullString
	ThursdayRoom  sql.NullString
	FridayRoom    sql.NullString
	SaturdayDates sql.NullString

	// Exams
	Partial1Date sql.NullTime
	Partial1Time sql.NullString
	Partial1Room sql.NullString

	Partial2Date sql.NullTime
	Partial2Time sql.NullString
	Partial2Room sql.NullString

	Final1Date    sql.NullTime
	Final1Time    sql.NullString
	Final1Room    sql.NullString
	Final1RevDate sql.NullTime
	Final1RevTime sql.NullString

	Final2Date    sql.NullTime
	Final2Time    sql.NullString
	Final2Room    sql.NullString
	Final2RevDate sql.NullTime
	Final2RevTime sql.NullString

	// Committee
	CommitteePresident sql.NullString
	CommitteeMember1   sql.NullString
	CommitteeMember2   sql.NullString
}
