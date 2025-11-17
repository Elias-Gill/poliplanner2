package parser

import (
	"testing"
)

var (
	metadataDir        = "../../testdata/excelparser/metadata"
	testCareerCode     = "example_metadata"
	testSubjectName    = "Algebra Lineal"
	normalizedSubject  = "algebra lineal"
	expectedSemester   = 2
	expectedCredits    = 5
)

func TestMetadataLoader_LoadsCareerCodes(t *testing.T) {
	loader := NewSubjectMetadataLoader(metadataDir)

	careerCodes := loader.GetAllCareerCodes()
	if len(careerCodes) == 0 {
		t.Fatal("No career codes loaded from metadata directory")
	}
	t.Logf("Loaded career codes: %v", careerCodes)
}

func TestMetadataLoader_FindSubjectByName(t *testing.T) {
	loader := NewSubjectMetadataLoader(metadataDir)

	metadata, err := loader.FindSubjectByName(testCareerCode, testSubjectName)
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
	loader := NewSubjectMetadataLoader(metadataDir)

	metadata, err := loader.FindSubjectByName(testCareerCode, "Álgebra Lineal (*)")
	if err != nil {
		t.Fatal("Failed to find normalized subject")
	}

	if metadata.Name != normalizedSubject {
		t.Errorf("Expected '%s' after normalization, got '%s'", normalizedSubject, metadata.Name)
	}
}

func TestMetadataLoader_DashedNames_FirstPart(t *testing.T) {
	loader := NewSubjectMetadataLoader(metadataDir)

	metadata, err := loader.FindSubjectByName(testCareerCode, "Cálculo I - Avanzado")
	if err != nil {
		t.Fatal("Failed to find subject with dash (first part)")
	}

	if metadata.Name != "calculo i" {
		t.Errorf("Expected 'calculo i' for dashed name, got '%s'", metadata.Name)
	}
}

func TestMetadataLoader_DashedNames_SecondPart(t *testing.T) {
    loader := NewSubjectMetadataLoader(metadataDir)

    metadata, err := loader.FindSubjectByName(testCareerCode, "Avanzado - Técnicas de Organización y metodos")
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
	loader := NewSubjectMetadataLoader(metadataDir)

	initialHits := loader.CacheHits

	// First search
	_, err := loader.FindSubjectByName(testCareerCode, testSubjectName)
	if err != nil {
		t.Fatal("Failed to find subject")
	}

	// Second search - should hit cache
	_, err = loader.FindSubjectByName(testCareerCode, testSubjectName)
	if err != nil {
		t.Fatal("Failed to find cached subject")
	}

	finalHits := loader.CacheHits
	if finalHits <= initialHits {
		t.Errorf("Expected cache hits to increase, got initial: %d, final: %d", initialHits, finalHits)
	}
}

func TestMetadataLoader_NonExistentSubject(t *testing.T) {
	loader := NewSubjectMetadataLoader(metadataDir)

	metadata, err := loader.FindSubjectByName(testCareerCode, "Materia Inexistente Que No Existe")
	if err == nil || metadata != nil {
		t.Error("Expected error for non-existent subject")
	}
}

func TestMetadataLoader_CaseAndAccentNormalization(t *testing.T) {
	loader := NewSubjectMetadataLoader(metadataDir)

	metadata, err := loader.FindSubjectByName(testCareerCode, "BASE de Datós I")
	if err != nil {
		t.Fatal("Failed to find subject with case and accent normalization")
	}

	if metadata.Name != "base de datos i" {
		t.Errorf("Expected 'base de datos i', got '%s'", metadata.Name)
	}
}

func TestMetadataLoader_EmptySubjectName(t *testing.T) {
	loader := NewSubjectMetadataLoader(metadataDir)

	metadata, err := loader.FindSubjectByName(testCareerCode, "")
	if err == nil || metadata != nil {
		t.Error("Expected error for empty subject name")
	}
}

func TestMetadataLoader_GetAllSubjects(t *testing.T) {
	loader := NewSubjectMetadataLoader(metadataDir)

	subjects := loader.GetSubjectsForCareer(testCareerCode)
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

func TestNormalizeName(t *testing.T) {
	loader := NewSubjectMetadataLoader(metadataDir)

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
		result := loader.normalizeName(test.input)
		if result != test.expected {
			t.Errorf("normalizeName(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}
