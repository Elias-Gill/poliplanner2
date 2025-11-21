package service

import "github.com/elias-gill/poliplanner2/internal/db/store"

var (
	userStorer           store.UserStorer
	sheetVersionStorer   store.SheetVersionStorer
	careerStorer         store.CareerStorer
	subjectStorer        store.SubjectStorer
	scheduleStorer       store.ScheduleStorer
	scheduleDetailStorer store.ScheduleDetailStorer
)

func InitializeServices(
	_userStorer store.UserStorer,
	_sheetVersionStorer store.SheetVersionStorer,
	_careerStorer store.CareerStorer,
	_subjectStorer store.SubjectStorer,
	_scheduleStorer store.ScheduleStorer,
	_scheduleDetailStorer store.ScheduleDetailStorer,
) {
	userStorer = _userStorer
	sheetVersionStorer = _sheetVersionStorer
	careerStorer = _careerStorer
	subjectStorer = _subjectStorer
	scheduleStorer = _scheduleStorer
	scheduleDetailStorer = _scheduleDetailStorer
}
