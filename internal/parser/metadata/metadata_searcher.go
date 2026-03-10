package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/elias-gill/poliplanner2/internal/config"
	log "github.com/elias-gill/poliplanner2/logger"
)

// ================================
// ======== Data Structures =======
// ================================

type AcademicPlanData struct {
	Name     string `json:"name"`
	Semester int    `json:"semester"`
	Credits  int    `json:"credits"`
}

type AcademicPlan struct {
	CareerCode string             `json:"career_code"`
	CareerName string             `json:"career_name"`
	Subjects   []AcademicPlanData `json:"subjects"`
}

type AcademicPlanLoader struct {
	metadataDir string
	careerCode  string
	subjects    []AcademicPlanData

	// Simple two-element cache for recent searches
	cachedName1     string
	cachedMetadata1 *AcademicPlanData
	cachedName2     string
	cachedMetadata2 *AcademicPlanData
	CacheHits       int
}

// ================================
// =        Public API            =
// ================================

func NewAcademicPlanLoader(careerCode string) (*AcademicPlanLoader, error) {
	metadataDir := path.Join(config.Get().Paths.BaseDir, "internal", "excel", "metadata", "curriculums")
	log.Debug("Creating subject metadata loader", "metadata_dir", metadataDir, "career_code", careerCode)

	loader := &AcademicPlanLoader{
		metadataDir: metadataDir,
		careerCode:  strings.ToLower(careerCode),
	}

	err := loader.loadSubjects()
	if err != nil {
		return nil, err
	}

	return loader, nil
}

func (loader *AcademicPlanLoader) FindSubject(subjectName string) (*AcademicPlanData, error) {
	if subjectName == "" {
		return nil, fmt.Errorf("subject name cannot be empty")
	}

	// Check fast cache comparing prefixes
	if loader.cachedName1 != "" && loader.matchesCache(subjectName, loader.cachedName1) {
		loader.CacheHits++
		// log.Debug("Cache hit on primary cache entry", "cache_hits", loader.CacheHits)
		return loader.cachedMetadata1, nil
	}
	if loader.cachedName2 != "" && loader.matchesCache(subjectName, loader.cachedName2) {
		loader.CacheHits++
		loader.swapCacheEntries()
		// log.Debug("Cache hit on secondary cache entry", "cache_hits", loader.CacheHits)
		return loader.cachedMetadata2, nil
	}

	// Not in cache, perform linear search by splitting the string when there is a '-'
	// e.g: "Electiva 1 - Materia completamente aburrida y al pedo"
	dashIndex := strings.Index(subjectName, "-")
	var part string
	if dashIndex > 0 {
		part = subjectName[:dashIndex]
	} else {
		part = subjectName
	}

	normalized := loader.normalizeName(part)
	found := loader.searchAcadamicData(normalized)

	// If not found with first part, try with second part
	if found == nil && dashIndex > 0 {
		secondPart := subjectName[dashIndex+1:]
		normalized = loader.normalizeName(secondPart)
		found = loader.searchAcadamicData(normalized)
	}

	loader.updateCache(subjectName, found)

	if found == nil {
		log.Warn("Subject metadata not found", "career", loader.careerCode, "subject", subjectName)
		return nil, fmt.Errorf("cannot find subject metadata for subject: %s", subjectName)
	}

	log.Debug("Subject found successfully",
		"subject", found.Name,
		"semester", found.Semester,
		"credits", found.Credits)
	return found, nil
}

// Just intended for debug
func (loader *AcademicPlanLoader) GetSubjectsList() []AcademicPlanData {
	return loader.subjects
}

// =====================================
// =        Private methods            =
// =====================================

func (loader *AcademicPlanLoader) loadSubjects() error {
	log.Debug("Loading subjects from file", "career", loader.careerCode)

	filePath := filepath.Join(loader.metadataDir, fmt.Sprintf("%s.json", loader.careerCode))
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	var careerSubjects AcademicPlan
	if err := json.Unmarshal(data, &careerSubjects); err != nil {
		return fmt.Errorf("error parsing JSON file %s: %v", filePath, err)
	}

	loader.subjects = careerSubjects.Subjects

	log.Debug("Successfully parsed career subjects",
		"career", careerSubjects.CareerCode,
		"career_name", careerSubjects.CareerName,
		"subjects_count", len(loader.subjects))

	return nil
}

func (loader *AcademicPlanLoader) searchAcadamicData(normalizedName string) *AcademicPlanData {
	for _, subject := range loader.subjects {
		if subject.Name == normalizedName {
			return &subject
		}
	}
	return nil
}

func (loader *AcademicPlanLoader) matchesCache(current, cached string) bool {
	return strings.HasPrefix(cached, current)
}

func (loader *AcademicPlanLoader) updateCache(name string, meta *AcademicPlanData) {
	log.Debug("Updating search cache", "name", name, "found", meta != nil)
	loader.cachedName2 = loader.cachedName1
	loader.cachedMetadata2 = loader.cachedMetadata1
	loader.cachedName1 = name
	loader.cachedMetadata1 = meta
}

func (loader *AcademicPlanLoader) swapCacheEntries() {
	log.Debug("Swapping cache entries")
	tempName := loader.cachedName1
	tempMeta := loader.cachedMetadata1
	loader.cachedName1 = loader.cachedName2
	loader.cachedMetadata1 = loader.cachedMetadata2
	loader.cachedName2 = tempName
	loader.cachedMetadata2 = tempMeta
}

func (loader *AcademicPlanLoader) normalizeName(raw string) string {
	if raw == "" {
		return ""
	}

	var sb strings.Builder
	lastWasSpace := false

	for _, r := range raw {
		// Ignore special chars
		if r == '*' || r == '(' || r == ')' {
			continue
		}

		// Remove accents
		r = loader.removeAccent(unicode.ToLower(r))
		if r == 0 {
			continue
		}

		// Normalize double spaces
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

	// finally trim leading and trilling spaces
	result := strings.TrimSpace(sb.String())

	log.Debug("Normalized subject name", "original", raw, "normalized", result)
	return result
}

func (loader *AcademicPlanLoader) removeAccent(c rune) rune {
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
		if c > 127 {
			return c
		}
		return c
	}
}
