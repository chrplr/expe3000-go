package engine

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidateExperiment checks if the experiment logical flow is correct and required files exist.
func ValidateExperiment(exp *Experiment, stimuliDir string) []error {
	var errors []error

	if len(exp.Stimuli) == 0 {
		errors = append(errors, fmt.Errorf("experiment has no valid stimuli"))
		return errors
	}

	for i, stim := range exp.Stimuli {
		if stim.DurationMS == 0 {
			errors = append(errors, fmt.Errorf("stimulus %d (at %dms) has zero duration", i+1, stim.TimestampMS))
		}

		if i > 0 {
			if stim.TimestampMS < exp.Stimuli[i-1].TimestampMS {
				errors = append(errors, fmt.Errorf("stimulus %d (at %dms) is out of order (previous at %dms)", i+1, stim.TimestampMS, exp.Stimuli[i-1].TimestampMS))
			}
		}

		if stim.Type == StimImage || stim.Type == StimSound || stim.Type == StimStream {
			for _, path := range stim.FilePaths {
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
		}
	}

	return errors
}
