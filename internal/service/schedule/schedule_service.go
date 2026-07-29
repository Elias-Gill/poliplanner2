package schedule

import (
	"context"
	"errors"
	"fmt"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/internal/model/schedule"
	"github.com/elias-gill/poliplanner2/internal/model/user"
	schedRepository "github.com/elias-gill/poliplanner2/internal/repository/schedule"
	"github.com/elias-gill/poliplanner2/logger"
)

var (
	ErrPermissionDenied  = errors.New("User has no permission")
	ErrTitleNotAvailable = errors.New("This title is already in use")
)

type ScheduleService struct {
	scheduleRepository schedRepository.ScheduleRepository
}

func New(scheduleRepo schedRepository.ScheduleRepository) *ScheduleService {
	return &ScheduleService{
		scheduleRepository: scheduleRepo,
	}
}

// ListUserSchedules returns all schedules for a user
func (s ScheduleService) ListUserSchedules(ctx context.Context, userID user.UserID) ([]schedule.ScheduleSummaryView, error) {
	logger.Debug("ListUserSchedules called", "userID", userID)
	sched, err := s.scheduleRepository.ListByUserID(ctx, userID)
	if err != nil {
		logger.Debug("cannot list user schedules", "userID", userID, "error", err)
		return nil, err
	}
	logger.Debug("ListUserSchedules successful", "userID", userID, "count", len(sched))
	return sched, nil
}

// ListUserSchedules returns the details for the dashboard view of a given schedule
func (s ScheduleService) GetScheduleOverview(ctx context.Context, userID user.UserID, scheduleID schedule.ScheduleID) (*schedule.StudentScheduleView, error) {
	logger.Debug("GetSchedule called", "userID", userID, "scheduleID", scheduleID)
	sche, err := s.scheduleRepository.GetDetailsByID(ctx, scheduleID)
	if err != nil {
		logger.Debug("cannot get schedule details", "scheduleID", scheduleID, "error", err)
		return nil, err
	}

	if sche.Owner != userID {
		logger.Debug("permission denied for schedule", "scheduleID", scheduleID, "userID", userID)
		return nil, ErrPermissionDenied
	}

	// Map info into our view models
	view := &schedule.StudentScheduleView{
		Weekly: s.buildWeeklySchedule(sche.Courses),
		Exams:  s.extractExams(sche.Courses),
		Info:   s.buildCoursesInfo(sche.Courses),
	}

	logger.Debug("GetSchedule successful", "scheduleID", scheduleID, "userID", userID)
	return view, nil
}

// Save persists a schedule and returns its ID
func (s ScheduleService) CreateSchedule(ctx context.Context, userID user.UserID, title string, courseIDs []academic.CourseID) (schedule.ScheduleID, error) {
	logger.Debug("CreateSchedule called", "title", title, "owner", userID)

	uID := user.UserID(userID)

	available, err := s.TitleIsAvailable(ctx, user.UserID(userID), title)
	if err != nil {
		logger.Error("cannot check title existence", "error", err)
		return -1, err
	}
	if !available {
		return -1, ErrTitleNotAvailable
	}

	sche, err := schedule.NewSchedule(uID, title, courseIDs)
	if err != nil {
		logger.Error("cannot create schedule entity", "error", err)
		return -1, err
	}

	id, err := s.scheduleRepository.Save(ctx, *sche)
	if err != nil {
		logger.Error("cannot save schedule", "error", err)
		return -1, err
	}

	return id, nil
}

func (s ScheduleService) Delete(ctx context.Context, userID user.UserID, scheduleID schedule.ScheduleID) error {
	logger.Debug("Delete schedule called", "userID", userID, "scheduleID", scheduleID)
	sche, err := s.scheduleRepository.GetDetailsByID(ctx, scheduleID)
	if err != nil {
		logger.Debug("cannot get schedule details for deletion", "scheduleID", scheduleID, "error", err)
		return err
	}

	if sche.Owner != userID {
		logger.Debug("permission denied for schedule", "scheduleID", scheduleID, "userID", userID)
		return ErrPermissionDenied
	}

	err = s.scheduleRepository.Delete(ctx, scheduleID)
	if err != nil {
		return err
	}

	logger.Debug("Schedule deletion successful", "scheduleID", scheduleID, "userID", userID)
	return nil
}

// TitleIsAvailable checks if the user has a schedule with the same title
func (s ScheduleService) TitleIsAvailable(ctx context.Context, userID user.UserID, title string) (bool, error) {
	logger.Debug("TitleIsAvailable called", "userID", userID, "title", title)
	list, err := s.ListUserSchedules(ctx, userID)
	if err != nil {
		logger.Error("cannot check title existence", "userID", userID, "title", title, "error", err)
		return false, err
	}

	for _, entry := range list {
		if entry.Title == title {
			logger.Debug("title already exists", "userID", userID, "title", title)
			return false, nil
		}
	}

	logger.Debug("title available", "userID", userID, "title", title)
	return true, nil
}

// 1. Armado de la grilla semanal para el dashboard
func (s ScheduleService) buildWeeklySchedule(courses []academic.CourseSummaryView) schedule.WeekScheduleView {
	var weekly schedule.WeekScheduleView

	for _, course := range courses {
		for _, session := range course.Schedules {
			classInfo := schedule.ClassSlotView{
				Course: course.Name,
				Room:   session.Room,
				Time:   session.Time,
			}

			// Map slot to specific day
			switch session.Day {
			case 1:
				weekly.Monday = append(weekly.Monday, classInfo)
			case 2:
				weekly.Tuesday = append(weekly.Tuesday, classInfo)
			case 3:
				weekly.Wednesday = append(weekly.Wednesday, classInfo)
			case 4:
				weekly.Thursday = append(weekly.Thursday, classInfo)
			case 5:
				weekly.Friday = append(weekly.Friday, classInfo)
			case 6:
				weekly.Saturday = append(weekly.Saturday, classInfo)
			}
		}
	}

	return weekly
}

// 2. Extracción y ordenamiento cronológico de exámenes (si aplican a las materias)
func (s ScheduleService) extractExams(courses []academic.CourseSummaryView) schedule.ExamMapView {
	var examMap schedule.ExamMapView

	for _, course := range courses {
		for _, exam := range course.Exams {
			slot := schedule.ExamSlotView{
				CourseName: course.Name,
				Room:       exam.Room,
			}

			if exam.HasDate() {
				if exam.HasHour() {
					// Formato para mostrar al usuario en el HTML
					slot.Date = exam.Date().Format("02/01/2006 - 15:04hs")
					// Formato estándar ISO para FullCalendar en JS
					slot.ISODate = exam.Date().Format("2006-01-02T15:04:00")
				} else {
					slot.Date = exam.Date().Format("02/01/2006")
					slot.ISODate = exam.Date().Format("2006-01-02")
				}
			}

			if exam.HasRevisionDate() {
				if exam.HasHour() {
					slot.Revision = exam.Revision().Format("02/01/2006 - 15:04hs")
					slot.ISORevision = exam.Revision().Format("2006-01-02T15:04:00")
				} else {
					slot.Revision = exam.Revision().Format("02/01/2006")
					slot.ISORevision = exam.Revision().Format("2006-01-02")
				}
			}

			// Agrupación según el tipo e instancia del examen
			switch exam.Type {
			case academic.ExamPartial:
				switch exam.Instance {
				case academic.Instance1:
					examMap.Partial1 = append(examMap.Partial1, slot)
				case academic.Instance2:
					examMap.Partial2 = append(examMap.Partial2, slot)
				}
			case academic.ExamFinal:
				switch exam.Instance {
				case academic.Instance1:
					examMap.Final1 = append(examMap.Final1, slot)
				case academic.Instance2:
					examMap.Final2 = append(examMap.Final2, slot)
				}
			}
		}
	}

	return examMap
}

func (s ScheduleService) buildCoursesInfo(courses []academic.CourseSummaryView) []schedule.CourseDetailView {
	infoList := make([]schedule.CourseDetailView, 0, len(courses))

	for _, course := range courses {
		// Mapeo del listado de docentes a TeacherContactView
		teachers := make([]schedule.TeacherContactView, 0, len(course.Teachers))
		for _, t := range course.Teachers {
			name := fmt.Sprintf("%s %s", t.FirstName, t.LastName)
			if t.Title != "" {
				name = fmt.Sprintf("%s %s", t.Title, name)
			}

			teachers = append(teachers, schedule.TeacherContactView{
				Name:  name,
				Email: t.Email,
			})
		}

		// Construcción de la vista detallada de la materia
		infoList = append(infoList, schedule.CourseDetailView{
			Name:               course.Name,
			Section:            course.Section,
			Shift:              course.Shift,
			Type:               course.Type,
			Teachers:           teachers,
			SaturdayDates:      course.SaturdayDates,
			CommitteePresident: course.Committee.President,
			CommitteeMember1:   course.Committee.Member1,
			CommitteeMember2:   course.Committee.Member2,
		})
	}

	return infoList
}
