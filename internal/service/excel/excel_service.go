package excel

import (
	"context"
	"fmt"
	"runtime"

	"time"

	"github.com/elias-gill/poliplanner2/internal/config/timezone"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/parser"
	"github.com/elias-gill/poliplanner2/internal/infrastructure/source"
	academicModel "github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/internal/model/excel"
	"github.com/elias-gill/poliplanner2/internal/repository"
	academicRepo "github.com/elias-gill/poliplanner2/internal/repository/academic"
	excelRepo "github.com/elias-gill/poliplanner2/internal/repository/excel"
	academicService "github.com/elias-gill/poliplanner2/internal/service/academic"
	metaServices "github.com/elias-gill/poliplanner2/internal/service/metadata"
	"github.com/elias-gill/poliplanner2/logger"
)

type ExcelService struct {
	excelRepository      excelRepo.ExcelRepository
	courseRepository     academicRepo.CourseRepository
	teacherRepository    academicRepo.TeacherRepository
	curriculumRepository academicRepo.CurriculumRepository
	periodRepository     academicRepo.PeriodRepository
	subjectRepository    academicRepo.SubjectRepository
	careerRepository     academicRepo.CareerRepository

	txManager repository.TxManager

	// --- External services ---

	periodService *academicService.PeriodService

	// --- Paths injected from the composition root ---

	layoutsDir  string
	metadataDir string
}

func NewExcelService(
	excelRepo excelRepo.ExcelRepository,
	courseRepo academicRepo.CourseRepository,
	teacherRepo academicRepo.TeacherRepository,
	curriculumRepo academicRepo.CurriculumRepository,
	periodRepo academicRepo.PeriodRepository,
	subjectRepo academicRepo.SubjectRepository,
	careerRepo academicRepo.CareerRepository,
	txManager repository.TxManager,
	periodService *academicService.PeriodService,
	layoutsDir string,
	metadataDir string,
) *ExcelService {
	return &ExcelService{
		excelRepository:      excelRepo,
		courseRepository:     courseRepo,
		teacherRepository:    teacherRepo,
		curriculumRepository: curriculumRepo,
		periodRepository:     periodRepo,
		subjectRepository:    subjectRepo,
		careerRepository:     careerRepo,
		txManager:            txManager,
		periodService:        periodService,
		layoutsDir:           layoutsDir,
		metadataDir:          metadataDir,
	}
}

// ListVersions returns the successfully parsed versions of a source type.
func (e ExcelService) ListVersions(ctx context.Context, kind excel.SourceType) ([]*excel.SheetVersion, error) {
	return e.excelRepository.ListVersions(ctx, kind)
}

// ListAudit returns the parse attempts of a source type, latest first.
func (e ExcelService) ListAudit(ctx context.Context, kind excel.SourceType, limit int) ([]*excel.ParseAudit, error) {
	return e.excelRepository.ListAudit(ctx, kind, limit)
}

// IsSourceUpToDate reports whether a version of the same source (matched by name
// or url) with an equal or newer source date was already parsed successfully.
func (e ExcelService) IsSourceUpToDate(
	ctx context.Context,
	kind excel.SourceType,
	name, url string,
	sourceDate time.Time,
) (bool, error) {
	return e.excelRepository.IsSourceUpToDate(ctx, kind, name, url, sourceDate)
}

// PersistScheduleSource parses and persists a schedule source, recording both
// the parse audit and the resulting version.
func (e ExcelService) PersistScheduleSource(ctx context.Context, src source.ScheduleSource) error {
	startedAt := time.Now().In(timezone.ParaguayTZ)

	periodID, sheetCount, parseErr := e.parseScheduleSource(ctx, src)

	finishedAt := time.Now().In(timezone.ParaguayTZ)
	meta := src.Metadata()

	audit := &excel.ParseAudit{
		SourceType:   excel.SourceTypeSchedule,
		Name:         meta.Name,
		URL:          meta.URI,
		SourceDate:   meta.Date,
		StartedAt:    startedAt,
		FinishedAt:   finishedAt,
		Succeeded:    parseErr == nil,
		ParsedSheets: sheetCount,
	}

	if parseErr == nil {
		versionID, err := e.excelRepository.SaveVersion(ctx, &excel.SheetVersion{
			PeriodID:     periodID,
			SourceType:   excel.SourceTypeSchedule,
			Name:         meta.Name,
			URL:          meta.URI,
			SourceDate:   meta.Date,
			ParsedAt:     finishedAt,
			ParsedSheets: sheetCount,
		})
		if err != nil {
			return fmt.Errorf("failed to save excel version (parse error: %v): %w", parseErr, err)
		}

		audit.VersionID = &versionID
	} else {
		audit.Error = parseErr.Error()
	}

	if err := e.excelRepository.SaveAudit(ctx, audit); err != nil {
		return fmt.Errorf("failed to save parse audit (parse error: %v): %w", parseErr, err)
	}

	if parseErr != nil {
		return fmt.Errorf("excel persistence transaction failed: %w", parseErr)
	}

	return nil
}

func (e ExcelService) parseScheduleSource(ctx context.Context, source source.ScheduleSource) (academicModel.PeriodID, int, error) {
	content, err := source.Content(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot open Excel source: %w", err)
	}
	defer content.Close()

	p, err := parser.NewScheduleParser(content, e.layoutsDir)
	if err != nil {
		return 0, 0, fmt.Errorf("cannot initialize excel parser: %w", err)
	}
	defer p.Close()

	// Upsert the period based on the provided source metadata
	periodID, err := e.periodRepository.Upsert(ctx, academicModel.Period{
		Year:     source.Metadata().Date.Year(),
		Semester: academicModel.YearSemester(source.Metadata().Semester),
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to upsert period: %w", err)
	}

	sheetCount := 0

	txErr := e.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		for {
			sheet, err := p.ParseNextSheet()
			if err != nil {
				return fmt.Errorf("error parsing sheet: %w", err)
			}
			if sheet == nil {
				break
			}

			logger.Info("Sheet parsing succesfull", "name", sheet)

			career := buildCareerFromDTO(sheet.Career)

			metadataService, err := metaServices.NewMetadataService(career.Code, e.metadataDir)
			if err != nil {
				return fmt.Errorf("error while loading metadata: %w", err)
			}

			metadataService.EnrichCareer(&career)

			careerID, err := e.careerRepository.Upsert(ctx, career)
			if err != nil {
				return fmt.Errorf("failed to upsert career '%s': %w", career.Code, err)
			}

			for _, data := range sheet.Subjects {
				// Load and enrich subject with known metadata
				sub := buildSubject(data)
				metadataService.EnrichSubject(&sub)

				subjectID, err := e.subjectRepository.Upsert(ctx, sub)
				if err != nil {
					return fmt.Errorf("failed to upsert subject '%s': %w", sub.Name, err)
				}

				// Load and enrich curriculum with known metadata
				curriculum := buildCurriculum(data)
				metadataService.EnrichCurriculum(sub, &curriculum)

				curriculumID, err := e.curriculumRepository.Upsert(ctx, academicRepo.CurriculumSaveParams{
					SubjectID:  subjectID,
					CareerID:   careerID,
					Curriculum: curriculum,
				})
				if err != nil {
					return fmt.Errorf("failed to upsert curriculum for subject '%s': %w", sub.Name, err)
				}

				// Teachers persistence
				teachers := buildTeachers(data.Teachers, data.TeacherCount)

				var teacherIDs []academicModel.TeacherID
				for _, t := range teachers {
					teacherID, err := e.teacherRepository.Upsert(ctx, t)
					if err != nil {
						return fmt.Errorf("failed to upsert teacher '%s': %w", t.FirstName, err)
					}
					teacherIDs = append(teacherIDs, teacherID)
				}

				// Build and persist course
				course := buildOfferingFromDTO(data)

				courseID, err := e.courseRepository.Upsert(ctx, &academicRepo.CourseSaveParams{
					Name:          course.Name,
					Type:          course.Type,
					Section:       course.Section,
					Shift:         course.Shift,
					Period:        periodID,
					Curriculum:    curriculumID,
					SaturdayDates: course.SaturdayDates,
					Comitee:       course.Comittee,
				})
				if err != nil {
					return fmt.Errorf("failed to upsert course '%s': %w", course.Name, err)
				}

				// Assign exams, schedule and teachers to the new course
				if err := e.courseRepository.AssignExams(ctx, courseID, course.Exams); err != nil {
					return fmt.Errorf("failed to assign exams to course '%s': %w", course.Name, err)
				}

				if err := e.courseRepository.AssignSchedule(ctx, courseID, course.Schedule); err != nil {
					return fmt.Errorf("failed to assign schedule to course '%s': %w", course.Name, err)
				}

				if err := e.courseRepository.AssignTeachers(ctx, courseID, teacherIDs); err != nil {
					return fmt.Errorf("failed to assign teachers to course '%s': %w", course.Name, err)
				}
			}

			sheetCount++

			// Force memmory cleaning
			sheet.Subjects = nil
			runtime.GC()

			logger.Info("Persisted succesfully")
		}

		return nil
	})

	if txErr != nil {
		return periodID, sheetCount, txErr
	}

	return periodID, sheetCount, nil
}

func (e ExcelService) PersistLabSource(ctx context.Context, source source.LabSource) error {
	logger.Debug("Persisting lab", "source", source.Metadata().Name)

	// TODO: Crear el repositorio para guardar y consultar los laboratorios

	// TODO: Implementar el servicio de laboratorios cuya API ya esta definida

	// TODO: Implementar la visualizacion en el frontend del dashboard

	return nil
}
