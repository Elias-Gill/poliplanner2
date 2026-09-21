package scraper

import "testing"

func TestScheduleFilter(t *testing.T) {
	tests := []struct {
		name string
		file string
		want bool
	}{
		{
			name: "schedule file is accepted",
			file: "Horario-de-clases-y-examenes-Segundo-Academico-2024-version-web-19122024.xlsx",
			want: true,
		},
		{
			name: "laboratory file is rejected",
			file: "Horario-de-Practicas-del-Laboratorio-Electronica-2025_1.xlsx",
			want: false,
		},
		{
			name: "unrelated file is rejected",
			file: "presupuesto-2025.xlsx",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scheduleFilter(tt.file); got != tt.want {
				t.Errorf("scheduleFilter(%q) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}

func TestLaboratoryFilter(t *testing.T) {
	tests := []struct {
		name string
		file string
		want bool
	}{
		{
			name: "laboratory file is accepted",
			file: "Horario-de-Practicas-del-Laboratorio-Electronica-2025_1.xlsx",
			want: true,
		},
		{
			name: "schedule file is rejected",
			file: "Horario-de-clases-y-examenes-Segundo-Academico-2024-version-web-19122024.xlsx",
			want: false,
		},
		{
			name: "unrelated file is rejected",
			file: "presupuesto-2025.xlsx",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := laboratoryFilter(tt.file); got != tt.want {
				t.Errorf("laboratoryFilter(%q) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}
