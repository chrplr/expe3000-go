package engine

import (
	"expe3000/internal/version"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/Zyko0/go-sdl3/img"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

type EventLogEntry struct {
	IntendedMS  uint64
	TimestampMS uint64
	Type        string
	Label       string
	StimulusRow []string
}

type EventLog struct {
	SubjectID         string
	CSVHeader         []string
	Entries           []EventLogEntry
	StartTime         string
	EndTime           string
	Completed         bool
	SDLVersion        string
	Platform          string
	Hostname          string
	Username          string
	VideoDriver       string
	AudioDriver       string
	Renderer          string
	DisplayMode       string
	LogicalResolution string
	Font              string
	FontSize          int
	CommandLine       string
}

func (l *EventLog) Log(intended, actual uint64, stype, label string, stimulusRow []string) {
	l.Entries = append(l.Entries, EventLogEntry{
		IntendedMS:  intended,
		TimestampMS: actual,
		Type:        stype,
		Label:       label,
		StimulusRow: stimulusRow,
	})
}

func isNumeric(s string) bool {
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}

func writeRow(f *os.File, row []string) {
	for i, field := range row {
		if i > 0 {
			f.WriteString("\t")
		}
		if isNumeric(field) {
			f.WriteString(field)
		} else {
			f.WriteString("\"")
			// Escape internal quotes
			f.WriteString(strings.ReplaceAll(field, "\"", "\"\""))
			f.WriteString("\"")
		}
	}
	f.WriteString("\n")
}

func (l *EventLog) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write metadata
	metadata := [][]string{
		{"# expe3000 version: " + version.Version + " (Go version)"},
		{"# Author: Christophe Pallier (christophe@pallier.org)"},
		{"# GitHub: https://github.com/chrplr/expe3000"},
		{"# SDL Version: " + l.SDLVersion},
		{"# Platform: " + l.Platform},
		{"# Hostname: " + l.Hostname},
		{"# Username: " + l.Username},
		{"# Video Driver: " + l.VideoDriver},
		{"# Audio Driver: " + l.AudioDriver},
		{"# Renderer: " + l.Renderer},
	}
	if l.DisplayMode != "" {
		metadata = append(metadata, []string{"# Display Mode: " + l.DisplayMode})
	}
	metadata = append(metadata, []string{"# Logical Resolution: " + l.LogicalResolution})

	fontName := l.Font
	if fontName == "" {
		fontName = "none"
	}
	metadata = append(metadata, []string{"# Font: " + fontName})
	metadata = append(metadata, []string{"# Font Size: " + strconv.Itoa(l.FontSize)})
	metadata = append(metadata, []string{"# Start Date: " + l.StartTime})
	metadata = append(metadata, []string{"# End Date: " + l.EndTime})

	completedStr := "Completed Normally"
	if !l.Completed {
		completedStr = "Aborted (ESC or Quit)"
	}
	metadata = append(metadata, []string{"# Completion Status: " + completedStr})
	metadata = append(metadata, []string{"# Command Line: " + l.CommandLine})

	for _, m := range metadata {
		writeRow(f, m)
	}

	// Write data header
	outputHdr := []string{"subject_id", "intended_ms", "actual_ms", "type", "label"}
	outputHdr = append(outputHdr, l.CSVHeader...)
	writeRow(f, outputHdr)

	// Write data entries
	for _, e := range l.Entries {
		row := []string{
			l.SubjectID,
			strconv.FormatUint(e.IntendedMS, 10),
			strconv.FormatUint(e.TimestampMS, 10),
			e.Type,
			e.Label,
		}
		if len(e.StimulusRow) > 0 {
			row = append(row, e.StimulusRow...)
		} else {
			for i := 0; i < len(l.CSVHeader); i++ {
				row = append(row, "")
			}
		}
		writeRow(f, row)
	}
	return nil
}

func DisplaySplash(renderer *sdl.Renderer, filePath string, screenW, screenH int, scaleFactor float32, bgColor sdl.Color) bool {
	if filePath == "" {
		return true
	}
	tex, err := img.LoadTexture(renderer, filePath)
	if err != nil {
		return true
	}
	defer tex.Destroy()

	tw, th, _ := tex.Size()
	dst := sdl.FRect{
		X: (float32(screenW) - tw*scaleFactor) / 2.0,
		Y: (float32(screenH) - th*scaleFactor) / 2.0,
		W: tw * scaleFactor,
		H: th * scaleFactor,
	}

	renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
	renderer.Clear()
	renderer.RenderTexture(tex, nil, &dst)
	renderer.Present()

	for {
		var event sdl.Event
		if err := sdl.WaitEvent(&event); err != nil {
			break
		}
		if event.Type == sdl.EVENT_QUIT {
			return false
		}
		if event.Type == sdl.EVENT_KEY_DOWN {
			break
		}
	}
	return true
}

func WaitForKeyPress(renderer *sdl.Renderer, font *ttf.Font, screenW, screenH int, textColor, bgColor sdl.Color) bool {
	if font == nil {
		return true
	}

	surf, err := font.RenderTextBlended("Press any key to start", textColor)
	if err != nil || surf == nil {
		return true
	}
	defer surf.Destroy()

	tex, err := renderer.CreateTextureFromSurface(surf)
	if err != nil {
		return true
	}
	defer tex.Destroy()

	dst := sdl.FRect{
		X: (float32(screenW) - float32(surf.W)) / 2.0,
		Y: (float32(screenH) - float32(surf.H)) / 2.0,
		W: float32(surf.W),
		H: float32(surf.H),
	}

	renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
	renderer.Clear()
	renderer.RenderTexture(tex, nil, &dst)
	renderer.Present()

	for {
		var event sdl.Event
		if err := sdl.WaitEvent(&event); err != nil {
			break
		}
		if event.Type == sdl.EVENT_QUIT {
			return false
		}
		if event.Type == sdl.EVENT_KEY_DOWN || event.Type == sdl.EVENT_MOUSE_BUTTON_DOWN {
			break
		}
	}
	return true
}

const CrossSize = 20

func drawFixationCross(renderer *sdl.Renderer, w, h int, color sdl.Color) {
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	mx, my := float32(w)/2, float32(h)/2
	renderer.RenderLine(mx-CrossSize, my, mx+CrossSize, my)
	renderer.RenderLine(mx, my-CrossSize, mx, my+CrossSize)
}

type experimentState struct {
	cfg       *Config
	exp       *Experiment
	resources []Resource
	renderer  *sdl.Renderer
	mixer     *AudioMixer
	log       *EventLog
	dlp       *DLPIO8G
	font      *ttf.Font

	stNS uint64 // Start time in NS
	ctNS uint64 // Current time in NS relative to start

	csIndex     int    // Current stimulus index
	activeVisual int    // Active visual stimulus index (-1 if none)
	visualEndNS  uint64 // End time for active visual stimulus in NS

	csidxSoundStream int    // Current sound index in a sound stream
	csvetSoundNS     uint64 // Next sound onset time in a sound stream in NS

	pulse2OffNS uint64 // Time to unset DLP line 2 in NS

	run     bool
	aborted bool

	laNS uint64 // Look-ahead in NS
}

func (s *experimentState) handleEvents() {
	for {
		var ev sdl.Event
		if !sdl.PollEvent(&ev) {
			break
		}
		switch ev.Type {
		case sdl.EVENT_QUIT:
			s.run = false
			s.aborted = true
		case sdl.EVENT_KEY_DOWN:
			ctNS := sdl.TicksNS() - s.stNS
			ctMS := ctNS / 1000000
			if ev.KeyboardEvent().Key == sdl.K_ESCAPE {
				s.run = false
				s.aborted = true
			} else {
				var activeRow []string
				if s.activeVisual != -1 {
					activeRow = s.exp.Stimuli[s.activeVisual].RawRow
				} else if s.csIndex > 0 && s.csIndex-1 < len(s.exp.Stimuli) {
					activeRow = s.exp.Stimuli[s.csIndex-1].RawRow
				}
				s.log.Log(ctMS, ctMS, "RESPONSE", ev.KeyboardEvent().Key.KeyName(), activeRow)
			}
		}
	}
}

func (s *experimentState) update() (bool, int) {
	s.ctNS = sdl.TicksNS() - s.stNS

	// Handle pulse-like DLP unsets
	if s.pulse2OffNS > 0 && s.ctNS >= s.pulse2OffNS {
		s.dlp.Unset("2")
		s.pulse2OffNS = 0
	}

	trig := false
	tidx := -1

	// Check for new stimulus onset
	if s.csIndex < len(s.exp.Stimuli) {
		stim := &s.exp.Stimuli[s.csIndex]
		onsetNS := stim.TimestampMS * 1000000

		if (s.ctNS + s.laNS) >= onsetNS {
			if (stim.Type == StimImage || stim.Type == StimText || stim.Type == StimBox || stim.Type == StimImageStream || stim.Type == StimTextStream) && len(s.resources[s.csIndex].Textures) > 0 {
				s.activeVisual = s.csIndex
				trig = true
				tidx = s.csIndex
				if stim.Type == StimImageStream || stim.Type == StimTextStream {
					totalDurNS := uint64(0)
					for i := 0; i < len(stim.FilePaths); i++ {
						totalDurNS += (stim.FrameDurations[i] + stim.FrameGaps[i]) * 1000000
					}
					s.visualEndNS = s.ctNS + totalDurNS
				} else {
					s.visualEndNS = s.ctNS + (stim.DurationMS * 1000000)
				}

				if s.dlp != nil {
					if stim.Type == StimImage || stim.Type == StimImageStream {
						s.dlp.Set("1")
					} else {
						s.dlp.Set("3") // TEXT and BOX on line 3
					}
				}
			} else if stim.Type == StimSound && len(s.resources[s.csIndex].Sounds) > 0 {
				if s.mixer.Play(&s.resources[s.csIndex].Sounds[0]) {
					s.log.Log(stim.TimestampMS, s.ctNS/1000000, "SOUND_ONSET", stim.FilePaths[0], stim.RawRow)
					if s.dlp != nil {
						s.dlp.Set("2")
						s.pulse2OffNS = s.ctNS + 5000000 // 5ms in NS
					}
				}
			} else if stim.Type == StimSoundStream && len(s.resources[s.csIndex].Sounds) > 0 {
				s.csidxSoundStream = 0
				if s.mixer.Play(&s.resources[s.csIndex].Sounds[0]) {
					s.log.Log(stim.TimestampMS, s.ctNS/1000000, "SOUND_STREAM_ONSET", strings.Join(stim.FilePaths, "~"), stim.RawRow)
					if s.dlp != nil {
						s.dlp.Set("2")
						s.pulse2OffNS = s.ctNS + 5000000
					}
				}
				s.csvetSoundNS = s.ctNS + (stim.FrameDurations[0]+stim.FrameGaps[0])*1000000
			}
			s.csIndex++
		}
	}

	// Handle Sound Streams
	if s.csidxSoundStream != -1 && s.csidxSoundStream+1 < len(s.resources[s.csIndex-1].Sounds) && s.ctNS >= s.csvetSoundNS {
		s.csidxSoundStream++
		stim := &s.exp.Stimuli[s.csIndex-1]
		if s.mixer.Play(&s.resources[s.csIndex-1].Sounds[s.csidxSoundStream]) {
			// Calculate intended MS based on cumulative previous frame durations and gaps
			intendedMS := stim.TimestampMS
			for i := 0; i < s.csidxSoundStream; i++ {
				intendedMS += stim.FrameDurations[i] + stim.FrameGaps[i]
			}
			s.log.Log(intendedMS, s.ctNS/1000000, "SOUND_STREAM_FRAME", stim.FilePaths[s.csidxSoundStream], stim.RawRow)
			if s.dlp != nil {
				s.dlp.Set("2")
				s.pulse2OffNS = s.ctNS + 5000000
			}
		}
		s.csvetSoundNS = s.ctNS + (stim.FrameDurations[s.csidxSoundStream]+stim.FrameGaps[s.csidxSoundStream])*1000000
		if s.csidxSoundStream == len(s.resources[s.csIndex-1].Sounds)-1 {
			s.csidxSoundStream = -1
		}
	}

	// Handle Visual Offsets
	if s.activeVisual != -1 && s.ctNS >= s.visualEndNS {
		stim := &s.exp.Stimuli[s.activeVisual]
		totalDurationMS := stim.TotalDuration()
		intendedOffMS := stim.TimestampMS + totalDurationMS
		label := strings.Join(stim.FilePaths, "~")
		stype := "IMAGE_OFFSET"
		switch stim.Type {
		case StimText:
			stype = "TEXT_OFFSET"
		case StimBox:
			stype = "BOX_OFFSET"
		case StimImageStream:
			stype = "IMAGE_STREAM_OFFSET"
		case StimTextStream:
			stype = "TEXT_STREAM_OFFSET"
		}
		s.log.Log(intendedOffMS, s.ctNS/1000000, stype, label, stim.RawRow)

		if s.dlp != nil {
			if stim.Type == StimImage || stim.Type == StimImageStream {
				s.dlp.Unset("1")
			} else {
				s.dlp.Unset("3")
			}
		}
		s.activeVisual = -1
	}

	// Check if finished
	if s.csIndex >= len(s.exp.Stimuli) && s.activeVisual == -1 && s.ctNS >= s.cfg.TotalDuration*1000000 {
		s.run = false
	}

	return trig, tidx
}

func (s *experimentState) render() {
	s.renderer.SetDrawColor(s.cfg.BGColor.R, s.cfg.BGColor.G, s.cfg.BGColor.B, s.cfg.BGColor.A)
	s.renderer.Clear()

	if s.activeVisual != -1 {
		r := &s.resources[s.activeVisual]
		stim := &s.exp.Stimuli[s.activeVisual]

		frameIdx := 0
		showBlank := false
		if stim.Type == StimImageStream || stim.Type == StimTextStream {
			totalDurationNS := uint64(0)
			for i := 0; i < len(stim.FilePaths); i++ {
				totalDurationNS += (stim.FrameDurations[i] + stim.FrameGaps[i]) * 1000000
			}
			elapsedNS := s.ctNS - (s.visualEndNS - totalDurationNS)
			
			// Find which frame we are in
			cumulNS := uint64(0)
			frameIdx = -1
			for i := 0; i < len(stim.FrameDurations); i++ {
				durNS := stim.FrameDurations[i] * 1000000
				gapNS := stim.FrameGaps[i] * 1000000
				if elapsedNS < cumulNS + durNS {
					frameIdx = i
					showBlank = false
					break
				}
				cumulNS += durNS
				if elapsedNS < cumulNS + gapNS {
					frameIdx = i
					showBlank = true
					break
				}
				cumulNS += gapNS
			}
			if frameIdx == -1 {
				frameIdx = len(r.Textures) - 1
			}
		}

		if !showBlank {
			tex := r.Textures[frameIdx]
			w := r.W[frameIdx]
			h := r.H[frameIdx]

			dr := sdl.FRect{
				X: (float32(s.cfg.ScreenWidth) - (w * s.cfg.ScaleFactor)) / 2.0,
				Y: (float32(s.cfg.ScreenHeight) - (h * s.cfg.ScaleFactor)) / 2.0,
				W: w * s.cfg.ScaleFactor,
				H: h * s.cfg.ScaleFactor,
			}
			s.renderer.RenderTexture(tex, nil, &dr)
		} else if s.cfg.UseFixation {
			drawFixationCross(s.renderer, s.cfg.ScreenWidth, s.cfg.ScreenHeight, s.cfg.FixationColor)
		}
	} else if s.cfg.UseFixation {
		drawFixationCross(s.renderer, s.cfg.ScreenWidth, s.cfg.ScreenHeight, s.cfg.FixationColor)
	}
	s.renderer.Present()
}

func RunExperiment(cfg *Config, exp *Experiment, resources []Resource, renderer *sdl.Renderer, mixer *AudioMixer, log *EventLog, dlp *DLPIO8G, font *ttf.Font) bool {
	prevGC := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(prevGC)

	rr := float32(60.0)
	win, _ := renderer.Window()
	display := sdl.GetDisplayForWindow(win)
	mode, err := display.CurrentDisplayMode()
	if err == nil && mode.RefreshRate > 0 {
		rr = mode.RefreshRate
	}
	fdNS := uint64(1000000000.0 / rr)

	laNS := fdNS / 2
	if cfg.VRR {
		laNS = 0 // In VRR we want to hit the timestamp exactly
	}

	state := &experimentState{
		cfg:              cfg,
		exp:              exp,
		resources:        resources,
		renderer:         renderer,
		mixer:            mixer,
		log:              log,
		dlp:              dlp,
		font:             font,
		csIndex:          0,
		activeVisual:     -1,
		csidxSoundStream: -1,
		run:              true,
		laNS:             laNS,
	}

	if cfg.VSync {
		// Sync start with a VBlank
		renderer.SetDrawColor(cfg.BGColor.R, cfg.BGColor.G, cfg.BGColor.B, cfg.BGColor.A)
		renderer.Clear()
		renderer.Present()
	}
	state.stNS = sdl.TicksNS()

	for state.run {
		state.handleEvents()
		if !state.run {
			break
		}

		// In VRR mode, if we are close to an onset, busy-wait to hit it exactly
		if cfg.VRR && state.csIndex < len(state.exp.Stimuli) {
			onsetNS := state.exp.Stimuli[state.csIndex].TimestampMS * 1000000
			ctNS := sdl.TicksNS() - state.stNS
			if ctNS < onsetNS && onsetNS-ctNS <= 2000000 { // If within 2ms, busy-wait
				for sdl.TicksNS()-state.stNS < onsetNS {
					// busy wait
				}
			}
		}

		trig, tidx := state.update()
		state.render()

		if trig {
			otNS := sdl.TicksNS() - state.stNS
			stim := &state.exp.Stimuli[tidx]
			label := strings.Join(stim.FilePaths, "~")
			stype := "IMAGE_ONSET"
			switch stim.Type {
			case StimText:
				stype = "TEXT_ONSET"
			case StimBox:
				stype = "BOX_ONSET"
			case StimImageStream:
				stype = "IMAGE_STREAM_ONSET"
			case StimTextStream:
				stype = "TEXT_STREAM_ONSET"
			}
			state.log.Log(stim.TimestampMS, otNS/1000000, stype, label, stim.RawRow)

			totalDurNS := uint64(stim.TotalDuration()) * 1000000
			state.visualEndNS = otNS + totalDurNS
		}

		if !cfg.VSync && !cfg.VRR {
			sdl.Delay(1)
		}
	}

	return !state.aborted
}
