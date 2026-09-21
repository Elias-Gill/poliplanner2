package commons

import "testing"

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"  Algebra   Lineal ", "ALGEBRA LINEAL"},
		{"T1", "T1"},
		{"t1", "T1"},
		{"", ""},
		{"   \t\n ", ""},
		{"seccion_1", "SECCION_1"},
	}

	for _, tt := range tests {
		if got := NormalizeKey(tt.in); got != tt.want {
			t.Errorf("NormalizeKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestScanLinesNRespectsLimit(t *testing.T) {
	var lines []string
	count := ScanLinesN("a\nb\nc\nd\ne", 3, func(_ int, line string) {
		lines = append(lines, line)
	})

	if count != 3 {
		t.Fatalf("ScanLinesN returned %d, want 3", count)
	}
	if len(lines) != 3 || lines[0] != "a" || lines[1] != "b" || lines[2] != "c" {
		t.Fatalf("ScanLinesN assigned %v, want [a b c]", lines)
	}
}

func TestScanLinesKeepsDefaultLimit(t *testing.T) {
	count := ScanLines("a\nb\nc\nd\ne", func(_ int, _ string) {})

	if count != 4 {
		t.Fatalf("ScanLines returned %d, want the default limit of 4", count)
	}
}

func TestScanLinesNSkipsBlankLines(t *testing.T) {
	var lines []string
	count := ScanLinesN("\n a \n\n b \n", 10, func(_ int, line string) {
		lines = append(lines, line)
	})

	if count != 2 || len(lines) != 2 || lines[0] != "a" || lines[1] != "b" {
		t.Fatalf("ScanLinesN assigned %v (count %d), want [a b]", lines, count)
	}
}
