package parser

import (
	"testing"
)

func TestNameNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Fisica 2", "fisica II"},
		{"calculo 7", "calculo VII"},
		{"Algebra 10", "algebra X"},
		{"programacion 1", "programacion I"},
		{"estadistica 20", "estadistica XX"},

		// Espacios y mayúsculas
		{"  Fisica   3  ", "fisica III"},
		{"CALCULO 5", "calculo V"},
		{"CáLCULO 5", "calculo V"},

		// No se debe normalizar
		{"fisica", "fisica"},
		{"fisica II", "fisica II"},
		{"fisica 0", "fisica 0"},
		{"fisica 21", "fisica 21"},

		{"Electiva 1 - Machine Learning", "electiva I"},
		{"electIVa 2 - quien sabe", "electiva II"},

		// Cosas con parentesis
		{"calculo V (variable vectorial)", "calculo V"},
		{"calculo V (*)", "calculo V"},
		{"calculo V (**)", "calculo V"},
	}

	for _, testCase := range tests {
		t.Run(testCase.input, func(t *testing.T) {
			dto := SubjectDTO{}
			dto.SetSubjectName(testCase.input)

			if dto.TentativeRealSubjectName != testCase.expected {
				t.Errorf(
					"SetSubjectName(%q) => %q, expected %q",
					testCase.input,
					dto.TentativeRealSubjectName,
					testCase.expected,
					)
			}
		})
	}
}
