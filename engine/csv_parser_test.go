package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExperimentWithHeadersAndComma(t *testing.T) {
	content := `onset_time,duration,type,stimuli,condition
0,1000,image,fixation.png,Baseline
1000,2000,sound,beep.wav,Test1
"3000",500,text,"Some Quoted Text",Test2
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	os.WriteFile(path, []byte(content), 0644)

	exp, err := LoadExperiment(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(exp.Stimuli) != 3 {
		t.Fatalf("expected 3 stimuli, got %d", len(exp.Stimuli))
	}

	if exp.Stimuli[0].TimestampMS != 0 {
		t.Errorf("expected timestamp 0, got %d", exp.Stimuli[0].TimestampMS)
	}

	if exp.Stimuli[2].TimestampMS != 3000 || exp.Stimuli[2].FilePath != "Some Quoted Text" {
		t.Errorf("failed to parse quoted and spaced fields: %+v", exp.Stimuli[2])
	}
}

func TestLoadExperimentSemicolon(t *testing.T) {
	content := `onset_time;duration;type;stimuli;condition
0;1000;image;fixation.png;Test
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	os.WriteFile(path, []byte(content), 0644)

	exp, err := LoadExperiment(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(exp.Stimuli) != 1 {
		t.Fatalf("expected 1 stimuli, got %d", len(exp.Stimuli))
	}

	if exp.Stimuli[0].FilePath != "fixation.png" {
		t.Errorf("failed to parse file path with semicolon, got %s", exp.Stimuli[0].FilePath)
	}
}

func TestLoadExperimentInvalid(t *testing.T) {
	content := `onset_time,duration,type,stimuli
foo,1000,image,fixation.png
0,bar,sound,beep.wav
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	os.WriteFile(path, []byte(content), 0644)

	_, err := LoadExperiment(path)
	if err == nil {
		t.Fatalf("expected error, got none")
	}
}
