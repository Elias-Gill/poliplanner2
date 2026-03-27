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
	Weekly courseOffering.CoursesScheduleView
	Exams  courseOffering.ExamsScheduleView
}

func NewDashboardRouter(
	scheduleService *schedule.ScheduleService,
	planService *academicPlan.AcademicPlanService,
) func(r chi.Router) {
	templateDir := path.Join(config.Get().Paths.BaseDir, "web", "templates", "pages", "dashboard")

	// Main page
	base := parseTemplateWithBaseLayout(path.Join(templateDir, "index.html"))

	dashboardTemplate := template.Must(template.Must(base.Clone()).ParseFiles(path.Join(templateDir, "overview.html")))
	calendarTemplate := template.Must(template.Must(base.Clone()).ParseFiles(path.Join(templateDir, "calendar.html")))

	return func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// extract last selected schedule cookie
			selected, present := getLatestSelectionCookie(r)

			// extract query param: ?id=...
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
			// 3. last schedule
			queryIDint, err := strconv.ParseInt(queryID, 10, 64)
			if queryID != "" && err == nil {
				selectedID = scheduleDomain.ScheduleID(queryIDint)
			} else if present {
				selectedID = scheduleDomain.ScheduleID(selected)
			} else if len(schedules) > 0 {
				selectedID = schedules[len(schedules)-1].ID
			}

			// TODO: fetch selected schedule data
			// FIX: error handling
			schedule, _ := scheduleService.GetSchedule(ctx, user.UserID(userID), selectedID)
			// if err != nil {
			// 	customRedirect(w, r, "/500")
			// 	return
			// }

			// TODO: load models into data
			weekly, _ := planService.ListCoursesSchedule(ctx, schedule.Courses)
			exams, _ := planService.ListCoursesExams(ctx, schedule.Courses)
			info, _ := planService.ListCoursesInfo(ctx, schedule.Courses)

			// TODO: render data models
			dashboardTemplate.Execute(w, overviewData{info, *weekly, exams})
		})

		r.Get("/calendar", func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			fakeExams := []courseOffering.ExamClass{
				{
					CourseName: "Matemática Discreta",
					Room:       "Aula 204",
					Date:       time.Date(2026, 4, 2, 0, 0, 0, 0, time.Local),
					Revision:   nil,
					Type:       courseOffering.ExamPartial,
					Instance:   courseOffering.Instance1,
				},
				{
					CourseName: "Programación II",
					Room:       "Laboratorio 3",
					Date:       time.Date(2026, 4, 5, 0, 0, 0, 0, time.Local),
					Revision:   &now,
					Type:       courseOffering.ExamPartial,
					Instance:   courseOffering.Instance2,
				},
				{
					CourseName: "Estructuras de Datos",
					Room:       "Laboratorio 1",
					Date:       now,
					Revision:   &now,
					Type:       courseOffering.ExamFinal,
					Instance:   courseOffering.Instance1,
				},
				{
					CourseName: "Álgebra Lineal",
					Room:       "Aula 210",
					Date:       time.Date(2026, 3, 30, 0, 0, 0, 0, time.Local),
					Revision:   nil,
					Type:       courseOffering.ExamFinal,
					Instance:   courseOffering.Instance2,
				},
			}
			calendarTemplate.Execute(w, fakeExams)
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
