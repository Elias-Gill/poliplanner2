package models

import "time"

// Subject represents the final domain model (equivalent to Java's Subject)
type Subject struct {
	SubjectID int64
	CareerID int64

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
	Monday    string
	Tuesday   string
	Wednesday string
	Thursday  string
	Friday    string
	Saturday  string

	// Classrooms
	MondayRoom    string
	TuesdayRoom   string
	WednesdayRoom string
	ThursdayRoom  string
	FridayRoom    string
	SaturdayDates string

	// Exams
	Partial1Date *time.Time
	Partial1Time string
	Partial1Room string

	Partial2Date *time.Time
	Partial2Time string
	Partial2Room string

	Final1Date    *time.Time
	Final1Time    string
	Final1Room    string
	Final1RevDate *time.Time
	Final1RevTime string

	Final2Date    *time.Time
	Final2Time    string
	Final2Room    string
	Final2RevDate *time.Time
	Final2RevTime string

	// Committee
	CommitteePresident string
	CommitteeMember1   string
	CommitteeMember2   string
}
