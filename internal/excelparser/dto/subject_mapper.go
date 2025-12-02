package dto

import (
	"database/sql"
	"time"

	"github.com/elias-gill/poliplanner2/internal/db/model"
)

func MapToSubject(dto SubjectDTO) *model.Subject {
	if (SubjectDTO{}) == dto {
		return nil
	}

	subject := &model.Subject{
		// General information
		Department:  dto.Department,
		SubjectName: dto.SubjectName,
		Semester:    dto.Semester,
		Section:     dto.Section,

		// Teacher information
		TeacherTitle:    dto.TeacherTitle,
		TeacherLastname: dto.TeacherLastName,
		TeacherName:     dto.TeacherName,
		TeacherEmail:    dto.TeacherEmail,

		// Committee (convertir string vacío a NullString)
		CommitteePresident: stringToNullString(dto.CommitteePresident),
		CommitteeMember1:   stringToNullString(dto.CommitteeMember1),
		CommitteeMember2:   stringToNullString(dto.CommitteeMember2),
	}

	// Weekly schedule (convertir string vacío a NullString)
	subject.Monday = stringToNullString(dto.Monday)
	subject.Tuesday = stringToNullString(dto.Tuesday)
	subject.Wednesday = stringToNullString(dto.Wednesday)
	subject.Thursday = stringToNullString(dto.Thursday)
	subject.Friday = stringToNullString(dto.Friday)
	subject.Saturday = stringToNullString(dto.Saturday)

	// Classrooms (convertir string vacío a NullString)
	subject.MondayRoom = stringToNullString(dto.MondayRoom)
	subject.TuesdayRoom = stringToNullString(dto.TuesdayRoom)
	subject.WednesdayRoom = stringToNullString(dto.WednesdayRoom)
	subject.ThursdayRoom = stringToNullString(dto.ThursdayRoom)
	subject.FridayRoom = stringToNullString(dto.FridayRoom)
	subject.SaturdayDates = stringToNullString(dto.SaturdayDates)

	// Partial 1 exams
	subject.Partial1Date = timePtrToNullTime(dto.Partial1Date)
	subject.Partial1Time = stringToNullString(dto.Partial1Time)
	subject.Partial1Room = stringToNullString(dto.Partial1Room)

	// Partial 2 exams
	subject.Partial2Date = timePtrToNullTime(dto.Partial2Date)
	subject.Partial2Time = stringToNullString(dto.Partial2Time)
	subject.Partial2Room = stringToNullString(dto.Partial2Room)

	// Final 1 exams
	subject.Final1Date = timePtrToNullTime(dto.Final1Date)
	subject.Final1Time = stringToNullString(dto.Final1Time)
	subject.Final1Room = stringToNullString(dto.Final1Room)
	subject.Final1RevDate = timePtrToNullTime(dto.Final1RevDate)
	subject.Final1RevTime = stringToNullString(dto.Final1RevTime)

	// Final 2 exams
	subject.Final2Date = timePtrToNullTime(dto.Final2Date)
	subject.Final2Time = stringToNullString(dto.Final2Time)
	subject.Final2Room = stringToNullString(dto.Final2Room)
	subject.Final2RevDate = timePtrToNullTime(dto.Final2RevDate)
	subject.Final2RevTime = stringToNullString(dto.Final2RevTime)

	return subject
}

// Helper function to convert string to sql.NullString
func stringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// Helper function to convert *time.Time to sql.NullTime
func timePtrToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}
