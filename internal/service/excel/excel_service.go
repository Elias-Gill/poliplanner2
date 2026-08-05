package excel

import (
	"context"
	"errors"
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
)

var (
	ErrNoSheetVersion = errors.New("No sheet version found")
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
	}
}

// GetLatestValidVersion lists the latest SUCCESFULLY parsed excel file version.
func (e ExcelService) GetLatestValidVersion(ctx context.Context) (*excel.SheetVersion, error) {
	versions, err := e.excelRepository.ListAllVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot find excel versions: %w", err)
	}

	// The repository returns an ordered list from latest to oldest.
	for _, v := range versions {
		// Check for the first correctly parsed version
		if v.Succeeded {
			return v, nil
		}
	}

	return nil, ErrNoSheetVersion
}

func (e ExcelService) ListVersions(ctx context.Context) ([]*excel.SheetVersion, error) {
	return e.excelRepository.ListAllVersions(ctx)
}

func (e ExcelService) PersistSource(ctx context.Context, source source.ScheduleSource) error {
	content, err := source.Content(ctx)
	if err != nil {
		return fmt.Errorf("cannot open Excel source: %w", err)
	}
	defer content.Close()

	p, err := parser.NewParser(content)
	if err != nil {
		return fmt.Errorf("cannot initialize excel parser: %w", err)
	}

	// Upsert the period based on the provided source metadata
	periodID, err := e.periodRepository.Upsert(ctx, academicModel.Period{
		Year:     source.Metadata().Date.Year(),
		Semester: academicModel.YearSemester(source.Metadata().Semester),
	})
	if err != nil {
		return fmt.Errorf("failed to upsert period: %w", err)
	}

	sheetCount := 0

	txErr := e.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		for p.NextSheet() {
			sheet, err := p.ParseCurrentSheet()
			if err != nil {
				if sheet != nil {
					return fmt.Errorf("error parsing sheet '%s': %w", sheet.Name, err)
				}
				return fmt.Errorf("error while parsing: %w", err)
			}

			career := buildCareerFromDTO(sheet.Name)

			metadataService, err := metaServices.NewMetadataService(career.Code)
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
					Comitee:       course.Comitee,
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
		}

		return nil
	})

	// Close parser after all sheets are processed
	p.Close()

	// REFACTOR: deberia de tener una tabla a parte de auditoria de parseo y una sola tabla para
	// las versiones de excel parseadas correctamente, asi puedo discernir entre lo correcto e
	// incorrecto sin problemas.

	// Prepare and save audit entries for the source persistence attemp
	var errMsg string
	if txErr != nil {
		errMsg = txErr.Error()
	}

	// Save audit entry, independently if the parsing and persistence process was succesfull or
	// not
	auditErr := e.excelRepository.SaveVersion(ctx, &excel.SheetVersion{
		PeriodID:     periodID,
		Name:         source.Metadata().Name,
		URL:          source.Metadata().URI,
		ParsedAt:     time.Now().In(timezone.ParaguayTZ),
		ParsedSheets: sheetCount,
		Succeeded:    txErr == nil,
		Error:        errMsg,
	})
	if auditErr != nil {
		return fmt.Errorf("failed to save excel version audit (original error: %v): %w", txErr, auditErr)
	}

	// Return error if the parsing and persist fails after saving the audit entry
	if txErr != nil {
		return fmt.Errorf("excel persistence transaction failed: %w", txErr)
	}

	// Correctly parsed and persisted
	return nil
}
