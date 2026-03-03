package engine

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateExperiment checks if the experiment logical flow is correct and required files exist.
func ValidateExperiment(exp *Experiment, stimuliDir string) []error {
	var errors []error

	if stimuliDir != "" {
		if info, err := os.Stat(stimuliDir); err != nil {
			if os.IsNotExist(err) {
				errors = append(errors, fmt.Errorf("stimuli directory does not exist: %s", stimuliDir))
			} else {
				errors = append(errors, fmt.Errorf("error checking stimuli directory: %v", err))
			}
		} else if !info.IsDir() {
			errors = append(errors, fmt.Errorf("stimuli directory path is a file, not a directory: %s", stimuliDir))
		}
	}

	if len(exp.Stimuli) == 0 {
		errors = append(errors, fmt.Errorf("experiment has no valid stimuli"))
		return errors
	}

	var lastVisualEnd uint64
	var lastVisualIndex int

	for i, stim := range exp.Stimuli {
		if stim.DurationMS == 0 {
			errors = append(errors, fmt.Errorf("stimulus %d (at %dms) has zero duration", i+1, stim.TimestampMS))
		}

		if i > 0 {
			if stim.TimestampMS < exp.Stimuli[i-1].TimestampMS {
				errors = append(errors, fmt.Errorf("stimulus %d (at %dms) is out of order (previous at %dms)", i+1, stim.TimestampMS, exp.Stimuli[i-1].TimestampMS))
			}
		}

		// Resource checks
		if stim.Type == StimImage || stim.Type == StimSound || stim.Type == StimImageStream || stim.Type == StimSoundStream {
			if len(stim.FilePaths) == 0 || (len(stim.FilePaths) == 1 && stim.FilePaths[0] == "") {
				errors = append(errors, fmt.Errorf("stimulus %d (at %dms): missing file path(s)", i+1, stim.TimestampMS))
			}

			for _, path := range stim.FilePaths {
				if path == "" {
					errors = append(errors, fmt.Errorf("stimulus %d (at %dms): empty file path in stream", i+1, stim.TimestampMS))
					continue
				}
				expectedPath := filepath.Join(stimuliDir, path)
				info, err := os.Stat(expectedPath)
				if err != nil {
					if os.IsNotExist(err) {
						errors = append(errors, fmt.Errorf("stimulus %d (at %dms): resource file not found: %s", i+1, stim.TimestampMS, expectedPath))
					} else {
						errors = append(errors, fmt.Errorf("stimulus %d (at %dms): error checking resource file: %v", i+1, stim.TimestampMS, err))
					}
				} else if info.IsDir() {
					errors = append(errors, fmt.Errorf("stimulus %d (at %dms): expected file, got directory: %s", i+1, stim.TimestampMS, expectedPath))
				}
			}
		} else if stim.Type == StimText || stim.Type == StimTextStream || stim.Type == StimBox {
			if len(stim.FilePaths) == 0 || (len(stim.FilePaths) == 1 && stim.FilePaths[0] == "") {
				errors = append(errors, fmt.Errorf("stimulus %d (at %dms): empty text string", i+1, stim.TimestampMS))
			}
		}

		// Overlap checks for visual stimuli (Image, Text, Box, Streams)
		isVisual := (stim.Type == StimImage || stim.Type == StimText || stim.Type == StimBox || stim.Type == StimImageStream || stim.Type == StimTextStream)
		if isVisual {
			var duration uint64
			if stim.Type == StimImageStream || stim.Type == StimTextStream {
				duration = stim.DurationMS * uint64(len(stim.FilePaths))
			} else {
				duration = stim.DurationMS
			}

			if stim.TimestampMS < lastVisualEnd {
				errors = append(errors, fmt.Errorf("stimulus %d (visual, at %dms) overlaps with previous visual stimulus %d (ends at %dms)", i+1, stim.TimestampMS, lastVisualIndex+1, lastVisualEnd))
			}

			lastVisualEnd = stim.TimestampMS + duration
			lastVisualIndex = i
		}
	}

	return errors
}
