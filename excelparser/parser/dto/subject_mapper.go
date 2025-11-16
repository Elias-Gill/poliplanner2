package dto

import (
	"github.com/elias-gill/poliplanner2/persistence"
)

func MapToSubject(dto SubjectDTO) *persistence.Subject {
	if (SubjectDTO{}) == dto {
		return nil
	}

	return &persistence.Subject{
		// General information
		Department:  dto.Department,
		SubjectName: dto.SubjectName,
		Semester:    dto.Semester,
		Section:     dto.Section,

		// Teacher information
		TeacherTitle:    dto.TeacherTitle,
		TeacherLastName: dto.TeacherLastName,
		TeacherName:     dto.TeacherName,
		TeacherEmail:    dto.TeacherEmail,

		// Weekly schedule
		Monday:    dto.Monday,
		Tuesday:   dto.Tuesday,
		Wednesday: dto.Wednesday,
		Thursday:  dto.Thursday,
		Friday:    dto.Friday,
		Saturday:  dto.Saturday,

		// Classrooms
		MondayRoom:    dto.MondayRoom,
		TuesdayRoom:   dto.TuesdayRoom,
		WednesdayRoom: dto.WednesdayRoom,
		ThursdayRoom:  dto.ThursdayRoom,
		FridayRoom:    dto.FridayRoom,
		SaturdayDates: dto.SaturdayDates,

		// Exam schedules
		Partial1Date: dto.Partial1Date,
		Partial1Time: dto.Partial1Time,
		Partial1Room: dto.Partial1Room,

		Partial2Date: dto.Partial2Date,
		Partial2Time: dto.Partial2Time,
		Partial2Room: dto.Partial2Room,

		Final1Date:    dto.Final1Date,
		Final1Time:    dto.Final1Time,
		Final1Room:    dto.Final1Room,
		Final1RevDate: dto.Final1RevDate,
		Final1RevTime: dto.Final1RevTime,

		Final2Date:    dto.Final2Date,
		Final2Time:    dto.Final2Time,
		Final2Room:    dto.Final2Room,
		Final2RevDate: dto.Final2RevDate,
		Final2RevTime: dto.Final2RevTime,

		// Committee
		CommitteePresident: dto.CommitteePresident,
		CommitteeMember1:   dto.CommitteeMember1,
		CommitteeMember2:   dto.CommitteeMember2,
	}
}
