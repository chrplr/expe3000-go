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

	if exp.Stimuli[2].TimestampMS != 3000 || exp.Stimuli[2].FilePaths[0] != "Some Quoted Text" {
		t.Errorf("failed to parse quoted and spaced fields: %+v", exp.Stimuli[2])
	}
}

func TestLoadExperimentStream(t *testing.T) {
	content := `onset_time,duration,type,stimuli
0,100,stream,img1.png~img2.png~img3.png
1000,200,text_stream,word1~word2~word3
2000,300,sound_stream,s1.wav~s2.wav~s3.wav
`
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	os.WriteFile(path, []byte(content), 0644)

	exp, err := LoadExperiment(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if exp.Stimuli[0].Type != StimImageStream || len(exp.Stimuli[0].FilePaths) != 3 {
		t.Errorf("failed to parse stream, got %+v", exp.Stimuli[0])
	}
	if exp.Stimuli[0].FilePaths[1] != "img2.png" {
		t.Errorf("expected img2.png, got %s", exp.Stimuli[0].FilePaths[1])
	}

	if exp.Stimuli[1].Type != StimTextStream || len(exp.Stimuli[1].FilePaths) != 3 {
		t.Errorf("failed to parse text_stream, got %+v", exp.Stimuli[1])
	}

	if exp.Stimuli[2].Type != StimSoundStream || len(exp.Stimuli[2].FilePaths) != 3 {
		t.Errorf("failed to parse sound_stream, got %+v", exp.Stimuli[2])
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

	if exp.Stimuli[0].FilePaths[0] != "fixation.png" {
		t.Errorf("failed to parse file path with semicolon, got %s", exp.Stimuli[0].FilePaths[0])
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
