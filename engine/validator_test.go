package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateExperiment(t *testing.T) {
	dir := t.TempDir()
	dummyFile := filepath.Join(dir, "image.png")
	os.WriteFile(dummyFile, []byte("fake"), 0644)

	exp := &Experiment{
		Stimuli: []Stimulus{
			{TimestampMS: 0, DurationMS: 1000, Type: StimImage, FilePath: "image.png"},
			{TimestampMS: 1000, DurationMS: 500, Type: StimImage, FilePath: "missing.png"},
			{TimestampMS: 500, DurationMS: 500, Type: StimText, FilePath: "Text"},
		},
	}

	errs := ValidateExperiment(exp, dir)

	if len(errs) != 2 {
		t.Fatalf("expected 2 validation errors, got %d: %v", len(errs), errs)
	}

	expectedMissing := "stimulus 2 (at 1000ms): resource file not found: " + filepath.Join(dir, "missing.png")
	if errs[0].Error() != expectedMissing {
		t.Errorf("unexpected error 1: %v", errs[0])
	}

	if errs[1].Error() != "stimulus 3 (at 500ms) is out of order (previous at 1000ms)" {
		t.Errorf("unexpected error 2: %v", errs[1])
	}
}
