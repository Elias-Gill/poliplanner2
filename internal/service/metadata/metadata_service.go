package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"unicode"

	"github.com/elias-gill/poliplanner2/internal/config"
	"github.com/elias-gill/poliplanner2/internal/model/academic"
	"github.com/elias-gill/poliplanner2/logger"
)

type subjects struct {
	Name     string `json:"name"`
	Semester int    `json:"semester"`
	Credits  int    `json:"credits"`
}

type emphasis struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type departmentsInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type careerInfo struct {
	CareerCode string     `json:"career_code"`
	CareerName string     `json:"career_name"`
	Emphasis   []emphasis `json:"emphasis"`
	Subjects   []subjects `json:"subjects"`
}

// ===================================
// =          Public API             =
// ===================================

// MetadataService provides data normalization and enrichment for academic domain
// models using static definitions. It maintains a two-level cache to optimize
// frequent string matching workflows during data processing.
type MetadataService struct {
	careerInfo      careerInfo
	departmentsInfo []departmentsInfo

	cachedName1     string
	cachedMetadata1 *subjects
	cachedName2     string
	cachedMetadata2 *subjects

	CacheHits     int
	hasCareerInfo bool // Guard flag to check if curriculum was successfully loaded
}

// NewMetadataService creates a service instance preloaded with the dataset
// corresponding to the requested career code. It propagates wrapped internal
// errors if configuration or filesystem read operations fail.
func NewMetadataService(careerCode string) (*MetadataService, error) {
	s := &MetadataService{}

	if err := s.loadData(careerCode); err != nil {
		return nil, fmt.Errorf("metadata service initialization failed: %w", err)
	}

	return s, nil
}

// EnrichCareer replaces or populates the career name with the canonical
// string defined in the static metadata configuration.
func (s *MetadataService) EnrichCareer(data *academic.Career) {
	if !s.hasCareerInfo || s.careerInfo.CareerName == "" {
		return
	}
	data.Name = s.careerInfo.CareerName
}

// EnrichSubject resolves the department code attached to the subject model
// and overwrites its descriptive name using the global departments mapping.
func (s *MetadataService) EnrichSubject(data *academic.Subject) {
	if data.Department.Code == "" {
		return
	}
	for i := range s.departmentsInfo {
		if s.departmentsInfo[i].Code == data.Department.Code {
			data.Department.Name = s.departmentsInfo[i].Name
			return
		}
	}
}

// EnrichCurriculum resolves the static semester definition for a given subject
// and assigns it to the target curriculum model.
func (s *MetadataService) EnrichCurriculum(subject academic.Subject, curriculum *academic.Curriculum) {
	if !s.hasCareerInfo {
		return
	}

	// Enrich emphases list
	for i := range curriculum.Emphases {
		for _, canonicalEmp := range s.careerInfo.Emphasis {
			if strings.EqualFold(canonicalEmp.Code, curriculum.Emphases[i].Code) ||
				strings.EqualFold(canonicalEmp.Name, curriculum.Emphases[i].Name) {

				curriculum.Emphases[i].Code = canonicalEmp.Code
				curriculum.Emphases[i].Name = canonicalEmp.Name
				break
			}
		}
	}

	if curriculum.Semester != 0 || subject.Name == "" {
		return
	}

	m, err := s.findSubjectInMetadata(subject.Name)
	if err != nil {
		return
	}

	curriculum.Semester = m.Semester
}

// ===================================
// =      Metadata search logic      =
// ===================================

func (s *MetadataService) findSubjectInMetadata(subjectName string) (*subjects, error) {
	if subjectName == "" {
		return nil, fmt.Errorf("subject name cannot be empty")
	}

	// Hit cache L1
	if s.cachedName1 != "" && strings.EqualFold(s.cachedName1, subjectName) {
		s.CacheHits++
		return s.cachedMetadata1, nil
	}
	// Hit cache L2
	if s.cachedName2 != "" && strings.EqualFold(s.cachedName2, subjectName) {
		s.CacheHits++
		s.swapCacheEntries()
		return s.cachedMetadata1, nil
	}

	dashIndex := strings.Index(subjectName, "-")
	var part string
	if dashIndex > 0 {
		part = subjectName[:dashIndex]
	} else {
		part = subjectName
	}

	normalized := s.normalizeName(part)
	found := s.searchAcademicData(normalized)

	if found == nil && dashIndex > 0 {
		secondPart := subjectName[dashIndex+1:]
		normalized = s.normalizeName(secondPart)
		found = s.searchAcademicData(normalized)
	}

	if found == nil {
		return nil, fmt.Errorf("subject metadata missing for: %s", subjectName)
	}

	s.updateCache(subjectName, found)

	return found, nil
}

func (s *MetadataService) searchAcademicData(normalizedName string) *subjects {
	for i := range s.careerInfo.Subjects {
		if s.careerInfo.Subjects[i].Name == normalizedName {
			return &s.careerInfo.Subjects[i]
		}
	}
	return nil
}

func (s *MetadataService) normalizeName(raw string) string {
	if raw == "" {
		return ""
	}

	var sb strings.Builder
	lastWasSpace := false

	for _, r := range raw {
		if r == '*' || r == '(' || r == ')' {
			continue
		}

		r = s.removeAccent(unicode.ToLower(r))
		if r == 0 {
			continue
		}

		if unicode.IsSpace(r) {
			if !lastWasSpace && sb.Len() > 0 {
				sb.WriteRune(' ')
				lastWasSpace = true
			}
		} else {
			sb.WriteRune(r)
			lastWasSpace = false
		}
	}

	return strings.TrimSpace(sb.String())
}

func (s *MetadataService) removeAccent(c rune) rune {
	switch c {
	case 'á':
		return 'a'
	case 'é':
		return 'e'
	case 'í':
		return 'i'
	case 'ó':
		return 'o'
	case 'ú':
		return 'u'
	case 'ü':
		return 'u'
	case 'ñ':
		return 'n'
	default:
		return c
	}
}

func (s *MetadataService) updateCache(name string, meta *subjects) {
	s.cachedName2 = s.cachedName1
	s.cachedMetadata2 = s.cachedMetadata1
	s.cachedName1 = name
	s.cachedMetadata1 = meta
}

func (s *MetadataService) swapCacheEntries() {
	s.cachedName1, s.cachedName2 = s.cachedName2, s.cachedName1
	s.cachedMetadata1, s.cachedMetadata2 = s.cachedMetadata2, s.cachedMetadata1
}

func (m *MetadataService) loadData(career string) error {
	career = strings.ToLower(career)
	basePath := config.Get().Paths.MetadataDir

	// Conditional load for the career curriculum file
	metaFile := path.Join(basePath, "curriculums", fmt.Sprintf("%s.json", career))
	data, err := os.ReadFile(metaFile)
	if err != nil {
		logger.Warn("Curriculum metadata file missing or unreadable during initialization", "file", metaFile)
		m.hasCareerInfo = false
	} else {
		if err := json.Unmarshal(data, &m.careerInfo); err != nil {
			return fmt.Errorf("failed to parse curriculum JSON %s: %w", metaFile, err)
		}
		m.hasCareerInfo = true
	}

	// Mandatory load for global departments dictionary
	depsFile := path.Join(basePath, "departments.json")
	depsData, err := os.ReadFile(depsFile)
	if err != nil {
		return fmt.Errorf("failed to read departments file %s: %w", depsFile, err)
	}

	if err := json.Unmarshal(depsData, &m.departmentsInfo); err != nil {
		return fmt.Errorf("failed to parse departments JSON %s: %w", depsFile, err)
	}

	return nil
}
