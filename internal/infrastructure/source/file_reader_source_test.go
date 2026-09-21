package source

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/elias-gill/poliplanner2/internal/model/academic"
)

func TestNewScheduleSourceFromReader(t *testing.T) {
	const body = "schedule-content"

	meta := SourceMetadata{
		Name:     "horario.xlsx",
		URI:      "upload://horario",
		Semester: academic.FirstSemester,
		Date:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	src := NewScheduleSourceFromReader(io.NopCloser(strings.NewReader(body)), meta)

	assertReaderSource(t, src, meta, body)
}

func TestNewLabSourceFromReader(t *testing.T) {
	const body = "laboratory-content"

	meta := SourceMetadata{
		Name:     "laboratorio.xlsx",
		URI:      "upload://laboratorio",
		Semester: academic.SecondSemester,
		Date:     time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	}

	src := NewLabSourceFromReader(io.NopCloser(strings.NewReader(body)), meta)

	assertReaderSource(t, src, meta, body)
}

func assertReaderSource(t *testing.T, src source, wantMeta SourceMetadata, wantBody string) {
	t.Helper()

	gotMeta := src.Metadata()
	if gotMeta.Name != wantMeta.Name {
		t.Errorf("Name = %q, want %q", gotMeta.Name, wantMeta.Name)
	}
	if gotMeta.URI != wantMeta.URI {
		t.Errorf("URI = %q, want %q", gotMeta.URI, wantMeta.URI)
	}
	if gotMeta.Semester != wantMeta.Semester {
		t.Errorf("Semester = %v, want %v", gotMeta.Semester, wantMeta.Semester)
	}
	if !gotMeta.Date.Equal(wantMeta.Date) {
		t.Errorf("Date = %v, want %v", gotMeta.Date, wantMeta.Date)
	}

	reader, err := src.Content(context.Background())
	if err != nil {
		t.Fatalf("Content: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if string(data) != wantBody {
		t.Errorf("Content = %q, want %q", string(data), wantBody)
	}
}
