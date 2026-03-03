package engine

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// detectDelimiter attempts to find if the file uses semicolons instead of commas as a delimiter.
func detectDelimiter(path string) (rune, error) {
	file, err := os.Open(path)
	if err != nil {
		return ',', err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		commaCount := strings.Count(line, ",")
		semiCount := strings.Count(line, ";")
		if semiCount > commaCount {
			return ';', nil
		}
	}
	return ',', nil
}

func LoadExperiment(path string) (*Experiment, error) {
	delimiter, err := detectDelimiter(path)
	if err != nil {
		return nil, fmt.Errorf("failed to detect delimiter: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = delimiter
	reader.TrimLeadingSpace = true
	// To handle lines dynamically without strict lengths, we could set FieldsPerRecord.
	// We'll leave it as default to let the reader figure it out, but handle errors below.

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("csv file is empty")
	}

	headers := records[0]
	idxOnset, idxDuration, idxType, idxStimuli := -1, -1, -1, -1

	for i, h := range headers {
		h = strings.ToLower(strings.TrimSpace(h))
		switch h {
		case "onset_time":
			idxOnset = i
		case "duration":
			idxDuration = i
		case "type":
			idxType = i
		case "stimuli":
			idxStimuli = i
		}
	}

	if idxOnset == -1 || idxDuration == -1 || idxType == -1 || idxStimuli == -1 {
		return nil, fmt.Errorf("csv missing required columns: 'onset_time', 'duration', 'type', 'stimuli'")
	}

	var stimuli []Stimulus
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) == 0 {
			continue
		}
		if len(record) <= idxOnset || len(record) <= idxDuration || len(record) <= idxType || len(record) <= idxStimuli {
			continue // skip malformed rows implicitly
		}

		timestampStr := strings.TrimSpace(record[idxOnset])
		timestamp, err := strconv.ParseUint(timestampStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid onset_time '%s': %v", i+1, record[idxOnset], err)
		}

		durationStr := strings.TrimSpace(record[idxDuration])
		duration, err := strconv.ParseUint(durationStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid duration '%s': %v", i+1, record[idxDuration], err)
		}

		var stype StimType
		var filePaths []string
		stimRaw := strings.TrimSpace(record[idxStimuli])

		switch strings.ToLower(strings.TrimSpace(record[idxType])) {
		case "image":
			stype = StimImage
			filePaths = []string{stimRaw}
		case "sound":
			stype = StimSound
			filePaths = []string{stimRaw}
		case "text":
			stype = StimText
			filePaths = []string{stimRaw}
		case "stream", "image_stream":
			stype = StimImageStream
			filePaths = strings.Split(stimRaw, "~")
			for i, p := range filePaths {
				filePaths[i] = strings.TrimSpace(p)
			}
		case "text_stream":
			stype = StimTextStream
			filePaths = strings.Split(stimRaw, "~")
			for i, p := range filePaths {
				filePaths[i] = strings.TrimSpace(p)
			}
		case "sound_stream":
			stype = StimSoundStream
			filePaths = strings.Split(stimRaw, "~")
			for i, p := range filePaths {
				filePaths[i] = strings.TrimSpace(p)
			}
		default:
			return nil, fmt.Errorf("line %d: unknown stimulus type: %s", i+1, record[idxType])
		}

		stimuli = append(stimuli, Stimulus{
			TimestampMS: timestamp,
			DurationMS:  duration,
			Type:        stype,
			FilePaths:   filePaths,
			RawRow:      record,
		})
	}

	return &Experiment{Header: headers, Stimuli: stimuli}, nil
}
