package metadata

import (
	"strings"
	"testing"
)

var (
	testCareerCode    = "test_metadata"
	testSubjectName   = "Algebra Lineal"
	normalizedSubject = "algebra lineal"
	expectedSemester  = 2
	expectedCredits   = 5
)

func TestMetadataLoader_FindSubjectByName(t *testing.T) {
	// Adaptado al nuevo constructor público
	service, err := NewMetadataService(testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	// Probamos el motor de búsqueda interno (que es privado pero accesible desde el mismo paquete)
	metadata, err := service.findSubjectInMetadata(testSubjectName)
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
	service, err := NewMetadataService(testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := service.findSubjectInMetadata("Álgebra Lineal (*)")
	if err != nil {
		t.Fatal("Failed to find normalized subject")
	}

	if metadata.Name != normalizedSubject {
		t.Errorf("Expected '%s' after normalization, got '%s'", normalizedSubject, metadata.Name)
	}
}

func TestMetadataLoader_DashedNames_FirstPart(t *testing.T) {
	service, err := NewMetadataService(testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := service.findSubjectInMetadata("Cálculo I - Avanzado")
	if err != nil {
		t.Fatal("Failed to find subject with dash (first part)")
	}

	if metadata.Name != "calculo i" {
		t.Errorf("Expected 'calculo i' for dashed name, got '%s'", metadata.Name)
	}
}

func TestMetadataLoader_DashedNames_SecondPart(t *testing.T) {
	service, err := NewMetadataService(testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := service.findSubjectInMetadata("Avanzado - Técnicas de Organización y metodos")
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
	service, err := NewMetadataService(testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	// 1. Primera búsqueda: Debe ser un Cache Miss
	_, err = service.findSubjectInMetadata(testSubjectName)
	if err != nil {
		t.Fatal("Failed to find subject")
	}

	if service.CacheHits != 0 {
		t.Errorf("Expected 0 cache hits on first search, got %d", service.CacheHits)
	}
	if !strings.EqualFold(service.cachedName1, testSubjectName) {
		t.Errorf("Expected cachedName1 to be %q, got %q", testSubjectName, service.cachedName1)
	}

	// 2. Segunda búsqueda: Debe ser un Cache Hit en L1
	_, err = service.findSubjectInMetadata(testSubjectName)
	if err != nil {
		t.Fatal("Failed to find cached subject")
	}

	if service.CacheHits != 1 {
		t.Errorf("Expected exactly 1 cache hit, got %d", service.CacheHits)
	}
	// Verificación de precisión: Sigue estando al frente de la caché
	if !strings.EqualFold(service.cachedName1, testSubjectName) {
		t.Errorf("Expected cachedName1 to remain %q, got %q", testSubjectName, service.cachedName1)
	}
}

func TestMetadataLoader_NonExistentSubject(t *testing.T) {
	service, err := NewMetadataService(testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := service.findSubjectInMetadata("Materia Inexistente Que No Existe")
	if err == nil || metadata != nil {
		t.Error("Expected error for non-existent subject")
	}
}

func TestMetadataLoader_CaseAndAccentNormalization(t *testing.T) {
	service, err := NewMetadataService(testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := service.findSubjectInMetadata("BASE de Datós I")
	if err != nil {
		t.Fatal("Failed to find subject with case and accent normalization")
	}

	if metadata.Name != "base de datos i" {
		t.Errorf("Expected 'base de datos i', got '%s'", metadata.Name)
	}
}

func TestMetadataLoader_EmptySubjectName(t *testing.T) {
	service, err := NewMetadataService(testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	metadata, err := service.findSubjectInMetadata("")
	if err == nil || metadata != nil {
		t.Error("Expected error for empty subject name")
	}
}

func TestMetadataLoader_GetAllSubjects(t *testing.T) {
	service, err := NewMetadataService(testCareerCode)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	// Adaptado: Accedemos directamente al slice interno de la estructura cargada
	subjects := service.careerInfo.Subjects
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
	service, err := NewMetadataService(testCareerCode)
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
		result := service.normalizeName(test.input)
		if result != test.expected {
			t.Errorf("normalizeName(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestMetadataLoader_InvalidCareerCode(t *testing.T) {
	service, _ := NewMetadataService("non_existent_career")
	if service.hasCareerInfo {
		t.Error("Expected false career info for invalid career code")
	}
}
