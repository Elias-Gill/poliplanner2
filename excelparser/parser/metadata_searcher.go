package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
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
	loader := &SubjectMetadataLoader{
		metadataDir:   metadataDir,
		subjectsCache: make(map[string][]SubjectMetadata),
	}

	loader.preloadAllMetadata()
	return loader
}

func (loader *SubjectMetadataLoader) FindSubjectByName(careerCode, subjectName string) (*SubjectMetadata, error) {
	if subjectName == "" {
		return nil, fmt.Errorf("Error: Subject without name\n")
	}

	// Check fast cache comparing prefixes
	if loader.cachedName1 != "" && loader.matchesCache(subjectName, loader.cachedName1) {
		loader.CacheHits++
		return loader.cachedMetadata1, nil
	}
	if loader.cachedName2 != "" && loader.matchesCache(subjectName, loader.cachedName2) {
		loader.CacheHits++
		loader.swapCacheEntries()
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
		return nil, fmt.Errorf("Error: Cannot found subject metadata for subject %s\n", subjectName)
	}
	return found, nil
}

func (loader *SubjectMetadataLoader) FindSubjectByExactName(careerCode, subjectName string) *SubjectMetadata {
	subjects := loader.GetSubjectsForCareer(careerCode)

	searchName := strings.ToLower(strings.TrimSpace(subjectName))

	for _, subject := range subjects {
		if strings.ToLower(strings.TrimSpace(subject.Name)) == searchName {
			return &subject
		}
	}

	return nil
}

func (loader *SubjectMetadataLoader) GetSubjectsForCareer(careerCode string) []SubjectMetadata {
	if subjects, exists := loader.subjectsCache[careerCode]; exists {
		return subjects
	}

	filePath := filepath.Join(loader.metadataDir, fmt.Sprintf("%s.json", careerCode))
	subjects, err := loader.loadSubjectsFromFile(filePath)
	if err != nil {
		fmt.Printf("Warning: No metadata found for career %s: %v\n", careerCode, err)
		return []SubjectMetadata{}
	}

	loader.subjectsCache[careerCode] = subjects
	return subjects
}

// NOTE: Intended for debug purposes
func (loader *SubjectMetadataLoader) GetAllCareerCodes() []string {
	var codes []string
	for code := range loader.subjectsCache {
		codes = append(codes, code)
	}
	return codes
}

// =====================================
// ======== Private methods ============
// =====================================

func (loader *SubjectMetadataLoader) preloadAllMetadata() {
	files, err := filepath.Glob(filepath.Join(loader.metadataDir, "*.json"))
	if err != nil {
		fmt.Printf("Warning: Error reading metadata directory: %v\n", err)
		return
	}

	for _, file := range files {
		careerCode := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))

		subjects, err := loader.loadSubjectsFromFile(file)
		if err != nil {
			fmt.Printf("Warning: Skipping invalid metadata file %s: %v\n", file, err)
			continue
		}

		loader.subjectsCache[careerCode] = subjects
		fmt.Printf("Loaded %d metadata subjects for career: %s\n", len(subjects), careerCode)
	}
}

func (loader *SubjectMetadataLoader) loadSubjectsFromFile(filePath string) ([]SubjectMetadata, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	var careerSubjects CareerSubjects
	if err := json.Unmarshal(data, &careerSubjects); err != nil {
		return nil, fmt.Errorf("error parsing JSON file %s: %v", filePath, err)
	}

	return careerSubjects.Subjects, nil
}

func (loader *SubjectMetadataLoader) searchMetadata(careerCode, normalizedName string) *SubjectMetadata {
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
	loader.cachedName2 = loader.cachedName1
	loader.cachedMetadata2 = loader.cachedMetadata1
	loader.cachedName1 = name
	loader.cachedMetadata1 = meta
}

func (loader *SubjectMetadataLoader) swapCacheEntries() {
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

	return strings.TrimSpace(sb.String())
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
			// FUTURE: For other accented characters, use simple mapping
			// In a more complete implementation, you could use golang.org/x/text/transform
			// or similar for full Unicode normalization
			return c
		}
		return c
	}
}
