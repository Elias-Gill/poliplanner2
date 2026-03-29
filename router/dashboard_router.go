package router

import (
	"context"
	"html/template"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/elias-gill/poliplanner2/internal/app/academicPlan"
	"github.com/elias-gill/poliplanner2/internal/app/schedule"
	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/domain/courseOffering"
	scheduleDomain "github.com/elias-gill/poliplanner2/internal/domain/schedule"
	"github.com/elias-gill/poliplanner2/internal/domain/user"
	"github.com/go-chi/chi/v5"
)

const latestSelectionCookie = "latestScheduleSelection"

type overviewData struct {
	Info   []courseOffering.CourseSummary
	Weekly *courseOffering.CoursesScheduleView
	Exams  *courseOffering.ExamsScheduleView
}

type dashboardPage struct {
	Mode       string
	SelectedID int64
	Data       any
	Schedules  []scheduleDomain.ScheduleSummary
}

func NewDashboardRouter(
	scheduleService *schedule.ScheduleService,
	planService *academicPlan.AcademicPlanService,
) func(r chi.Router) {
	templateDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages", "dashboard")

	base := parseTemplateWithBaseLayout(path.Join(templateDir, "index.html"))
	dashboardTemplate := template.Must(template.Must(base.Clone()).ParseFiles(path.Join(templateDir, "overview.html")))
	calendarTemplate := template.Must(template.Must(base.Clone()).ParseFiles(path.Join(templateDir, "calendar.html")))

	// External functions to serve data
	// FIX: error handling
	serveOverview := func(ctx context.Context, userID user.UserID, selectedID scheduleDomain.ScheduleID) (any, error) {
		schedule, err := scheduleService.GetSchedule(ctx, userID, selectedID)
		if err != nil {
			return nil, err
		}

		weekly, _ := planService.ListCoursesSchedule(ctx, schedule.Courses)
		exams, _ := planService.ListCoursesExams(ctx, schedule.Courses)
		info, _ := planService.ListCoursesInfo(ctx, schedule.Courses)

		return overviewData{info, weekly, exams}, nil
	}

	// FIX: error handling
	serveCalendar := func(ctx context.Context, userID user.UserID, selectedID scheduleDomain.ScheduleID) (any, error) {
		schedule, err := scheduleService.GetSchedule(ctx, userID, selectedID)
		if err != nil {
			return nil, err
		}
		exams, _ := planService.ListCoursesExams(ctx, schedule.Courses)
		return exams, nil
	}

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// extract mode query param
			mode := r.URL.Query().Get("mode")
			if mode != "calendar" {
				mode = "overview"
			}

			// extract last selected schedule cookie
			selected, present := getLatestSelectionCookie(r)
			queryID := r.URL.Query().Get("id")
			userID := extractUserID(r)

			ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*300)
			defer cancel()

			// List user schedules
			schedules, err := scheduleService.ListUserSchedules(ctx, user.UserID(userID))
			if err != nil {
				// FIX: diferenciar entre error de no hay horarios con el error de
				// servidor. Basicamente no deberia de fallar si es que el server de
				// autenticacion esta haciendo su trabajo.
				//
				// Deberia de hecho hechar de la sesion en caso de haber error o algo asi,
				// porque no es normal tampoco. Deberia de refactorear el tema de las sesiones
				// en un futuro cercano.
				customRedirect(w, r, "/500")
				return
			}

			var selectedID scheduleDomain.ScheduleID

			// priority:
			// 1. query param
			// 2. cookie
			// 3. last schedule on the list
			queryIDint, err := strconv.ParseInt(queryID, 10, 64)
			if queryID != "" && err == nil {
				selectedID = scheduleDomain.ScheduleID(queryIDint)
			} else if present {
				selectedID = scheduleDomain.ScheduleID(selected)
			} else if len(schedules) > 0 {
				selectedID = schedules[len(schedules)-1].ID
			}

			var data any
			if mode == "calendar" {
				data, _ = serveCalendar(ctx, user.UserID(userID), selectedID)
				calendarTemplate.Execute(w, dashboardPage{Mode: mode, SelectedID: int64(selectedID), Data: data, Schedules: schedules})
				return
			}

			data, _ = serveOverview(ctx, user.UserID(userID), selectedID)
			dashboardTemplate.Execute(w, dashboardPage{Mode: mode, SelectedID: int64(selectedID), Data: data, Schedules: schedules})
		})
	}
}

func getLatestSelectionCookie(r *http.Request) (int64, bool) {
	cookie, err := r.Cookie(latestSelectionCookie)
	if err != nil {
		return -1, false
	}

	id, err := strconv.ParseInt(cookie.String(), 10, 64)
	if err != nil || id < 1 {
		return -1, false
	}

	return id, true
}
