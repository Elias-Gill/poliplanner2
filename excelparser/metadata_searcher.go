package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	log "github.com/elias-gill/poliplanner2/logger"
)

// ================================
// ======== Data Structures =======
// ================================

type SubjectMetadata struct {
	Name     string `json:"name"`
	Semester int    `json:"semester"`
	Credits  int    `json:"credits"`
}

type CareerSubjects struct {
	CareerCode string            `json:"career_code"`
	CareerName string            `json:"career_name"`
	Subjects   []SubjectMetadata `json:"subjects"`
}

type SubjectMetadataLoader struct {
	metadataDir   string
	subjectsCache map[string][]SubjectMetadata

	// Simple two-element cache for recent searches
	cachedName1     string
	cachedMetadata1 *SubjectMetadata
	cachedName2     string
	cachedMetadata2 *SubjectMetadata
	CacheHits       int
}

// ================================
// ======== Public API ============
// ================================

func NewSubjectMetadataLoader(metadataDir string) *SubjectMetadataLoader {
	log.Logger.Debug("Creating subject metadata loader", "metadata_dir", metadataDir)

	loader := &SubjectMetadataLoader{
		metadataDir:   metadataDir,
		subjectsCache: make(map[string][]SubjectMetadata),
	}

	loader.preloadAllMetadata()
	return loader
}

func (loader *SubjectMetadataLoader) FindSubjectByName(careerCode, subjectName string) (*SubjectMetadata, error) {
	if subjectName == "" {
		return nil, fmt.Errorf("subject name cannot be empty")
	}

	log.Logger.Debug("Searching for subject", "career", careerCode, "subject", subjectName)

	// Check fast cache comparing prefixes
	if loader.cachedName1 != "" && loader.matchesCache(subjectName, loader.cachedName1) {
		loader.CacheHits++
		log.Logger.Debug("Cache hit on primary cache entry", "cache_hits", loader.CacheHits)
		return loader.cachedMetadata1, nil
	}
	if loader.cachedName2 != "" && loader.matchesCache(subjectName, loader.cachedName2) {
		loader.CacheHits++
		loader.swapCacheEntries()
		log.Logger.Debug("Cache hit on secondary cache entry", "cache_hits", loader.CacheHits)
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
	found := loader.searchMetadata(careerCode, normalized)

	// If not found with first part, try with second part
	if found == nil && dashIndex > 0 {
		secondPart := subjectName[dashIndex+1:]
		normalized = loader.normalizeName(secondPart)
		found = loader.searchMetadata(careerCode, normalized)
	}

	loader.updateCache(subjectName, found)

	if found == nil {
		log.Logger.Warn("Subject metadata not found", "career", careerCode, "subject", subjectName)
		return nil, fmt.Errorf("cannot find subject metadata for subject: %s", subjectName)
	}

	log.Logger.Debug("Subject found successfully",
		"subject", found.Name,
		"semester", found.Semester,
		"credits", found.Credits)
	return found, nil
}

func (loader *SubjectMetadataLoader) FindSubjectByExactName(careerCode, subjectName string) *SubjectMetadata {
	log.Logger.Debug("Searching for subject by exact name", "career", careerCode, "subject", subjectName)

	subjects := loader.GetSubjectsForCareer(careerCode)

	searchName := strings.ToLower(strings.TrimSpace(subjectName))

	for _, subject := range subjects {
		if strings.ToLower(strings.TrimSpace(subject.Name)) == searchName {
			log.Logger.Debug("Exact match found", "subject", subject.Name)
			return &subject
		}
	}

	log.Logger.Debug("No exact match found")
	return nil
}

func (loader *SubjectMetadataLoader) GetSubjectsForCareer(careerCode string) []SubjectMetadata {
	if subjects, exists := loader.subjectsCache[careerCode]; exists {
		log.Logger.Debug("Career subjects found in cache", "career", careerCode, "subjects_count", len(subjects))
		return subjects
	}

	log.Logger.Debug("Loading career subjects from file", "career", careerCode)
	filePath := filepath.Join(loader.metadataDir, fmt.Sprintf("%s.json", careerCode))
	subjects, err := loader.loadSubjectsFromFile(filePath)
	if err != nil {
		log.Logger.Warn("No metadata found for career", "career", careerCode, "error", err)
		return []SubjectMetadata{}
	}

	loader.subjectsCache[careerCode] = subjects
	log.Logger.Debug("Career subjects loaded and cached", "career", careerCode, "subjects_count", len(subjects))
	return subjects
}

// NOTE: Intended for debug purposes
func (loader *SubjectMetadataLoader) GetAllCareerCodes() []string {
	codes := make([]string, 0, len(loader.subjectsCache))
	for code := range loader.subjectsCache {
		codes = append(codes, code)
	}
	log.Logger.Debug("Retrieved all career codes", "count", len(codes))
	return codes
}

// =====================================
// ======== Private methods ============
// =====================================

func (loader *SubjectMetadataLoader) preloadAllMetadata() {
	log.Logger.Info("Preloading all metadata files", "directory", loader.metadataDir)

	files, err := filepath.Glob(filepath.Join(loader.metadataDir, "*.json"))
	if err != nil {
		log.Logger.Warn("Error reading metadata directory", "error", err)
		return
	}

	log.Logger.Debug("Found metadata files", "count", len(files))
	totalLoaded := 0

	for _, file := range files {
		careerCode := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

		subjects, err := loader.loadSubjectsFromFile(file)
		if err != nil {
			log.Logger.Warn("Skipping invalid metadata file", "file", file, "error", err)
			continue
		}

		loader.subjectsCache[careerCode] = subjects
		totalLoaded += len(subjects)
		log.Logger.Debug("Loaded metadata subjects",
			"career", careerCode,
			"subjects_count", len(subjects),
			"file", file)
	}

	log.Logger.Info("Metadata preloading completed",
		"careers_loaded", len(loader.subjectsCache),
		"total_subjects", totalLoaded)
}

func (loader *SubjectMetadataLoader) loadSubjectsFromFile(filePath string) ([]SubjectMetadata, error) {
	log.Logger.Debug("Loading subjects from file", "file", filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	var careerSubjects CareerSubjects
	if err := json.Unmarshal(data, &careerSubjects); err != nil {
		return nil, fmt.Errorf("error parsing JSON file %s: %v", filePath, err)
	}

	log.Logger.Debug("Successfully parsed career subjects",
		"career", careerSubjects.CareerCode,
		"career_name", careerSubjects.CareerName,
		"subjects_count", len(careerSubjects.Subjects))

	return careerSubjects.Subjects, nil
}

func (loader *SubjectMetadataLoader) searchMetadata(careerCode, normalizedName string) *SubjectMetadata {
	log.Logger.Debug("Searching in metadata", "career", careerCode, "normalized_name", normalizedName)

	subjects := loader.GetSubjectsForCareer(careerCode)

	for _, subject := range subjects {
		if subject.Name == normalizedName {
			return &subject
		}
	}
	return nil
}

func (loader *SubjectMetadataLoader) matchesCache(current, cached string) bool {
	return strings.HasPrefix(cached, current)
}

func (loader *SubjectMetadataLoader) updateCache(name string, meta *SubjectMetadata) {
	log.Logger.Debug("Updating search cache", "name", name, "found", meta != nil)
	loader.cachedName2 = loader.cachedName1
	loader.cachedMetadata2 = loader.cachedMetadata1
	loader.cachedName1 = name
	loader.cachedMetadata1 = meta
}

func (loader *SubjectMetadataLoader) swapCacheEntries() {
	log.Logger.Debug("Swapping cache entries")
	tempName := loader.cachedName1
	tempMeta := loader.cachedMetadata1
	loader.cachedName1 = loader.cachedName2
	loader.cachedMetadata1 = loader.cachedMetadata2
	loader.cachedName2 = tempName
	loader.cachedMetadata2 = tempMeta
}

func (loader *SubjectMetadataLoader) normalizeName(raw string) string {
	if raw == "" {
		return ""
	}

	var sb strings.Builder
	lastWasSpace := false

	for _, r := range raw {
		if r == '*' || r == '(' || r == ')' {
			continue
		}

		r = loader.removeAccent(unicode.ToLower(r))
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

	result := strings.TrimSpace(sb.String())
	log.Logger.Debug("Normalized subject name", "original", raw, "normalized", result)
	return result
}

func (loader *SubjectMetadataLoader) removeAccent(c rune) rune {
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
