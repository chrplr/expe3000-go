package engine

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// detectDelimiter attempts to find the most likely delimiter in a CSV file.
func detectDelimiter(path string) (rune, error) {
	file, err := os.Open(path)
	if err != nil {
		return ',', err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for i := 0; i < 5 && scanner.Scan(); i++ {
		lines = append(lines, scanner.Text())
	}

	if len(lines) == 0 {
		return ',', nil
	}

	delimiters := []rune{',', ';', '\t', '|'}
	bestDelim := delimiters[0]
	maxScore := -1

	for _, d := range delimiters {
		colCounts := make([]int, 0, len(lines))
		for _, line := range lines {
			r := csv.NewReader(strings.NewReader(line))
			r.Comma = d
			r.LazyQuotes = true
			rec, err := r.Read()
			if err == nil {
				colCounts = append(colCounts, len(rec))
			}
		}

		if len(colCounts) == 0 {
			continue
		}

		// Check consistency
		consistent := true
		for i := 1; i < len(colCounts); i++ {
			if colCounts[i] != colCounts[0] || colCounts[i] < 2 {
				consistent = false
				break
			}
		}

		score := 0
		if consistent && colCounts[0] >= 4 { // Minimum 4 columns for expe3000
			score = 1000 + colCounts[0]
		} else {
			for _, c := range colCounts {
				if c > 1 {
					score += c
				}
			}
		}

		if score > maxScore {
			maxScore = score
			bestDelim = d
		}
	}

	return bestDelim, nil
}

// parseStreamString parses strings like "image.png:200:50" or "image.png:200" or just "image.png"
func parseStreamString(s string, defaultDuration uint64) (string, uint64, uint64) {
	parts := strings.Split(s, ":")
	path := strings.TrimSpace(parts[0])
	duration := defaultDuration
	gap := uint64(0)

	if len(parts) >= 2 {
		d, err := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 64)
		if err == nil {
			duration = d
		}
	}
	if len(parts) >= 3 {
		g, err := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 64)
		if err == nil {
			gap = g
		}
	}
	return path, duration, gap
}

func LoadExperiment(path string) (*Experiment, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}

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
	reader.LazyQuotes = true

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
		var frameDurations []uint64
		var frameGaps []uint64
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
			rawPaths := strings.Split(stimRaw, "~")
			for _, p := range rawPaths {
				path, d, g := parseStreamString(p, duration)
				filePaths = append(filePaths, path)
				frameDurations = append(frameDurations, d)
				frameGaps = append(frameGaps, g)
			}
		case "text_stream":
			stype = StimTextStream
			rawPaths := strings.Split(stimRaw, "~")
			for _, p := range rawPaths {
				path, d, g := parseStreamString(p, duration)
				filePaths = append(filePaths, path)
				frameDurations = append(frameDurations, d)
				frameGaps = append(frameGaps, g)
			}
		case "sound_stream":
			stype = StimSoundStream
			rawPaths := strings.Split(stimRaw, "~")
			for _, p := range rawPaths {
				path, d, g := parseStreamString(p, duration)
				filePaths = append(filePaths, path)
				frameDurations = append(frameDurations, d)
				frameGaps = append(frameGaps, g)
			}
		case "box":
			stype = StimBox
			// Convert literal "\n" to actual newlines
			content := strings.ReplaceAll(stimRaw, "\\n", "\n")
			filePaths = []string{content}
		default:
			return nil, fmt.Errorf("line %d: unknown stimulus type: %s", i+1, record[idxType])
		}

		stimuli = append(stimuli, Stimulus{
			TimestampMS:    timestamp,
			DurationMS:     duration,
			Type:           stype,
			FilePaths:      filePaths,
			FrameDurations: frameDurations,
			FrameGaps:      frameGaps,
			RawRow:         record,
		})
	}

	return &Experiment{Header: headers, Stimuli: stimuli}, nil
}
