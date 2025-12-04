package service

import (
	"database/sql"

	"github.com/elias-gill/poliplanner2/internal/db/store"
)

var (
	db                   *sql.DB
	userStorer           store.UserStorer
	sheetVersionStorer   store.SheetVersionStorer
	careerStorer         store.CareerStorer
	subjectStorer        store.SubjectStorer
	scheduleStorer       store.ScheduleStorer
	scheduleDetailStorer store.ScheduleDetailStorer
)

func InitializeServices(
	conn *sql.DB,
	userStore store.UserStorer,
	sheetVersionStore store.SheetVersionStorer,
	careerStore store.CareerStorer,
	subjectStore store.SubjectStorer,
	scheduleStore store.ScheduleStorer,
	scheduleDetailStore store.ScheduleDetailStorer,
) {
	db = conn
	userStorer = userStore
	sheetVersionStorer = sheetVersionStore
	careerStorer = careerStore
	subjectStorer = subjectStore
	scheduleStorer = scheduleStore
	scheduleDetailStorer = scheduleDetailStore
}
