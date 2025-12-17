package parser

import (
	"testing"
)

var (
	metadataDir       = "../../testdata/excelparser/metadata"
	testCareerCode    = "example_metadata"
	testSubjectName   = "Algebra Lineal"
	normalizedSubject = "algebra lineal"
	expectedSemester  = 2
	expectedCredits   = 5
)

func TestMetadataLoader_FindSubjectByName(t *testing.T) {
	loader, err := NewSubjectMetadataLoader(metadataDir, testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := loader.FindSubjectByName(testSubjectName)
	if err != nil {
		t.Fatal("Failed to find subject")
	}

	if metadata.Name != normalizedSubject {
		t.Errorf("Expected '%s', got '%s'", normalizedSubject, metadata.Name)
	}
	if metadata.Semester != expectedSemester {
		t.Errorf("Expected semester %d, got %d", expectedSemester, metadata.Semester)
	}
	if metadata.Credits != expectedCredits {
		t.Errorf("Expected %d credits, got %d", expectedCredits, metadata.Credits)
	}
}

func TestMetadataLoader_NameNormalization(t *testing.T) {
	loader, err := NewSubjectMetadataLoader(metadataDir, testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := loader.FindSubjectByName("Álgebra Lineal (*)")
	if err != nil {
		t.Fatal("Failed to find normalized subject")
	}

	if metadata.Name != normalizedSubject {
		t.Errorf("Expected '%s' after normalization, got '%s'", normalizedSubject, metadata.Name)
	}
}

func TestMetadataLoader_DashedNames_FirstPart(t *testing.T) {
	loader, err := NewSubjectMetadataLoader(metadataDir, testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := loader.FindSubjectByName("Cálculo I - Avanzado")
	if err != nil {
		t.Fatal("Failed to find subject with dash (first part)")
	}

	if metadata.Name != "calculo i" {
		t.Errorf("Expected 'calculo i' for dashed name, got '%s'", metadata.Name)
	}
}

func TestMetadataLoader_DashedNames_SecondPart(t *testing.T) {
	loader, err := NewSubjectMetadataLoader(metadataDir, testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := loader.FindSubjectByName("Avanzado - Técnicas de Organización y metodos")
	if err != nil {
		t.Fatalf("Error finding subject: %v", err)
	}

	if metadata == nil {
		t.Fatal("Failed to find subject with dash (second part) - returned nil")
	}

	if metadata.Name != "tecnicas de organizacion y metodos" {
		t.Errorf("Expected 'tecnicas de organizacion y metodos' for second part, got '%s'", metadata.Name)
	}
}

func TestMetadataLoader_CacheFunctionality(t *testing.T) {
	loader, err := NewSubjectMetadataLoader(metadataDir, testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	initialHits := loader.CacheHits

	// First search
	_, err = loader.FindSubjectByName(testSubjectName)
	if err != nil {
		t.Fatal("Failed to find subject")
	}

	// Second search - should hit cache
	_, err = loader.FindSubjectByName(testSubjectName)
	if err != nil {
		t.Fatal("Failed to find cached subject")
	}

	finalHits := loader.CacheHits
	if finalHits <= initialHits {
		t.Errorf("Expected cache hits to increase, got initial: %d, final: %d", initialHits, finalHits)
	}
}

func TestMetadataLoader_NonExistentSubject(t *testing.T) {
	loader, err := NewSubjectMetadataLoader(metadataDir, testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := loader.FindSubjectByName("Materia Inexistente Que No Existe")
	if err == nil || metadata != nil {
		t.Error("Expected error for non-existent subject")
	}
}

func TestMetadataLoader_CaseAndAccentNormalization(t *testing.T) {
	loader, err := NewSubjectMetadataLoader(metadataDir, testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := loader.FindSubjectByName("BASE de Datós I")
	if err != nil {
		t.Fatal("Failed to find subject with case and accent normalization")
	}

	if metadata.Name != "base de datos i" {
		t.Errorf("Expected 'base de datos i', got '%s'", metadata.Name)
	}
}

func TestMetadataLoader_EmptySubjectName(t *testing.T) {
	loader, err := NewSubjectMetadataLoader(metadataDir, testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := loader.FindSubjectByName("")
	if err == nil || metadata != nil {
		t.Error("Expected error for empty subject name")
	}
}

func TestMetadataLoader_GetAllSubjects(t *testing.T) {
	loader, err := NewSubjectMetadataLoader(metadataDir, testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	subjects := loader.GetSubjects()
	if len(subjects) == 0 {
		t.Error("Expected subjects for career, got 0")
	}

	t.Logf("Found %d subjects for career %s", len(subjects), testCareerCode)
	for i, subject := range subjects {
		if i < 5 {
			t.Logf("  %d: %s (semester: %d, credits: %d)",
				i+1, subject.Name, subject.Semester, subject.Credits)
		}
	}
}

func (loader *SubjectMetadataLoader) normalizeNameForTest(input string) string {
	return loader.normalizeName(input)
}

func TestNormalizeName(t *testing.T) {
	loader, err := NewSubjectMetadataLoader(metadataDir, testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	normalizationTests := []struct {
		input    string
		expected string
	}{
		{"Álgebra Lineal", "algebra lineal"},
		{"Cálculo I (*)", "calculo i"},
		{"Técnicas (Avanzadas)", "tecnicas avanzadas"},
		{"Base de Datós I", "base de datos i"},
		{"Programación II (**)", "programacion ii"},
		{"  Espacios   Extra  ", "espacios extra"},
		{"", ""},
	}

	for _, test := range normalizationTests {
		result := loader.normalizeNameForTest(test.input)
		if result != test.expected {
			t.Errorf("normalizeName(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

// Verify constructor error when file does not exists
func TestMetadataLoader_InvalidCareerCode(t *testing.T) {
	loader, err := NewSubjectMetadataLoader(metadataDir, "non_existent_career")
	if err == nil {
		t.Error("Expected error for invalid career code")
	}
	if loader != nil {
		t.Error("Expected nil loader for invalid career code")
	}
}
