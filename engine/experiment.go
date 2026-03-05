package engine

import (
	"encoding/csv"
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

func (l *EventLog) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Comma = '\t'

	// Write metadata
	metadata := [][]string{
		{"# expe3000 version: " + version.Version + " (Go version)"},
		{"# Author: " + version.Author},
		{"# License: " + version.License},
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
		w.Write(m)
	}

	// Write data header
	outputHdr := []string{"subject_id", "intended_ms", "actual_ms", "type", "label"}
	outputHdr = append(outputHdr, l.CSVHeader...)
	w.Write(outputHdr)

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
		w.Write(row)
	}
	w.Flush()
	return w.Error()
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

const (
	msToNS      = 1000000
	dlpPulseNS  = 5 * msToNS
	vrrWaitNS   = 2 * msToNS
	gracePeriod = 500
)

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

	csIndex      int    // Current stimulus index
	activeVisual int    // Active visual stimulus index (-1 if none)
	visualStartNS uint64 // Start time for active visual stimulus in NS
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
			ctMS := ctNS / msToNS
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

func (s *experimentState) handleDLPPulses() {
	if s.pulse2OffNS > 0 && s.ctNS >= s.pulse2OffNS {
		if s.dlp != nil {
			s.dlp.Unset("2")
		}
		s.pulse2OffNS = 0
	}
}

func (s *experimentState) checkStimulusOnset() (bool, int) {
	if s.csIndex >= len(s.exp.Stimuli) {
		return false, -1
	}

	stim := &s.exp.Stimuli[s.csIndex]
	onsetNS := stim.TimestampMS * msToNS

	if (s.ctNS + s.laNS) >= onsetNS {
		trig := false
		tidx := s.csIndex

		switch stim.Type {
		case StimImage, StimText, StimBox, StimImageStream, StimTextStream, StimVideo:
			if (stim.Type == StimVideo && s.resources[s.csIndex].Video != nil) || len(s.resources[s.csIndex].Textures) > 0 {
				s.activeVisual = s.csIndex
				s.visualStartNS = s.ctNS
				
				durMS := stim.TotalDuration()
				if stim.Type == StimVideo && durMS == 0 {
					v := s.resources[s.csIndex].Video
					durMS = uint64(float64(v.Video.Header.FrameCount) / v.FPS * 1000.0)
				}
				s.visualEndNS = s.visualStartNS + durMS*msToNS
				trig = true
				if s.dlp != nil {
					if stim.Type == StimImage || stim.Type == StimImageStream || stim.Type == StimVideo {
						s.dlp.Set("1")
					} else {
						s.dlp.Set("3") // TEXT and BOX on line 3
					}
				}
			}
		case StimSound:
			if len(s.resources[s.csIndex].Sounds) > 0 {
				if s.mixer.Play(&s.resources[s.csIndex].Sounds[0]) {
					s.log.Log(stim.TimestampMS, s.ctNS/msToNS, "SOUND_ONSET", stim.FilePaths[0], stim.RawRow)
					if s.dlp != nil {
						s.dlp.Set("2")
						s.pulse2OffNS = s.ctNS + dlpPulseNS
					}
				}
			}
		case StimSoundStream:
			if len(s.resources[s.csIndex].Sounds) > 0 {
				s.csidxSoundStream = 0
				if s.mixer.Play(&s.resources[s.csIndex].Sounds[0]) {
					s.log.Log(stim.TimestampMS, s.ctNS/msToNS, "SOUND_STREAM_ONSET", strings.Join(stim.FilePaths, "~"), stim.RawRow)
					if s.dlp != nil {
						s.dlp.Set("2")
						s.pulse2OffNS = s.ctNS + dlpPulseNS
					}
				}
				s.csvetSoundNS = s.ctNS + (stim.FrameDurations[0]+stim.FrameGaps[0])*msToNS
			}
		}
		s.csIndex++
		return trig, tidx
	}
	return false, -1
}

func (s *experimentState) advanceSoundStream() {
	if s.csidxSoundStream == -1 {
		return
	}

	stim := &s.exp.Stimuli[s.csIndex-1]
	if s.csidxSoundStream+1 < len(s.resources[s.csIndex-1].Sounds) && s.ctNS >= s.csvetSoundNS {
		s.csidxSoundStream++
		if s.mixer.Play(&s.resources[s.csIndex-1].Sounds[s.csidxSoundStream]) {
			intendedMS := stim.TimestampMS
			for i := 0; i < s.csidxSoundStream; i++ {
				intendedMS += stim.FrameDurations[i] + stim.FrameGaps[i]
			}
			s.log.Log(intendedMS, s.ctNS/msToNS, "SOUND_STREAM_FRAME", stim.FilePaths[s.csidxSoundStream], stim.RawRow)
			if s.dlp != nil {
				s.dlp.Set("2")
				s.pulse2OffNS = s.ctNS + dlpPulseNS
			}
		}
		s.csvetSoundNS = s.ctNS + (stim.FrameDurations[s.csidxSoundStream]+stim.FrameGaps[s.csidxSoundStream])*msToNS
		if s.csidxSoundStream == len(s.resources[s.csIndex-1].Sounds)-1 {
			s.csidxSoundStream = -1
		}
	}
}

func (s *experimentState) checkVisualOffset() {
	if s.activeVisual != -1 && s.ctNS >= s.visualEndNS {
		stim := &s.exp.Stimuli[s.activeVisual]
		totalDurationMS := stim.TotalDuration()
		intendedOffMS := stim.TimestampMS + totalDurationMS
		label := strings.Join(stim.FilePaths, "~")
		stype := stim.Type.String() + "_OFFSET"

		s.log.Log(intendedOffMS, s.ctNS/msToNS, stype, label, stim.RawRow)

		if s.dlp != nil {
			if stim.Type == StimImage || stim.Type == StimImageStream {
				s.dlp.Unset("1")
			} else {
				s.dlp.Unset("3")
			}
		}
		s.activeVisual = -1
	}
}

func (s *experimentState) checkFinished() {
	if s.csIndex >= len(s.exp.Stimuli) && s.activeVisual == -1 && s.ctNS >= s.cfg.TotalDuration*msToNS {
		s.run = false
	}
}

func (s *experimentState) update() (bool, int) {
	s.ctNS = sdl.TicksNS() - s.stNS

	s.handleDLPPulses()
	trig, tidx := s.checkStimulusOnset()
	s.advanceSoundStream()
	s.checkVisualOffset()
	s.checkFinished()

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
		tex := (*sdl.Texture)(nil)
		w, h := float32(0), float32(0)

		if stim.Type == StimImageStream || stim.Type == StimTextStream {
			elapsedNS := s.ctNS - s.visualStartNS

			// Find which frame we are in
			cumulNS := uint64(0)
			frameIdx = -1
			for i := 0; i < len(stim.FrameDurations); i++ {
				durNS := stim.FrameDurations[i] * msToNS
				gapNS := stim.FrameGaps[i] * msToNS
				if elapsedNS < cumulNS+durNS {
					frameIdx = i
					showBlank = false
					break
				}
				cumulNS += durNS
				if elapsedNS < cumulNS+gapNS {
					frameIdx = i
					showBlank = true
					break
				}
				cumulNS += gapNS
			}
			if frameIdx == -1 {
				frameIdx = len(r.Textures) - 1
			}
			if frameIdx < len(r.Textures) {
				tex = r.Textures[frameIdx]
				w = r.W[frameIdx]
				h = r.H[frameIdx]
			}
		} else if stim.Type == StimVideo {
			v := r.Video
			if v != nil {
				elapsedNS := s.ctNS - s.visualStartNS
				targetFrame := int(float64(elapsedNS) / 1e9 * v.FPS)

				v.UpdateFrame(targetFrame)

				if v.LastFrame >= 0 {
					tex = v.Texture
					w = v.W
					h = v.H
				}
			}
		} else {
			if len(r.Textures) > 0 {
				tex = r.Textures[0]
				w = r.W[0]
				h = r.H[0]
			}
		}

		if tex != nil && !showBlank {
			dr := sdl.FRect{
				X: (float32(s.cfg.ScreenWidth) - (w * s.cfg.ScaleFactor)) / 2.0,
				Y: (float32(s.cfg.ScreenHeight) - (h * s.cfg.ScaleFactor)) / 2.0,
				W: w * s.cfg.ScaleFactor,
				H: h * s.cfg.ScaleFactor,
			}
			s.renderer.RenderTexture(tex, nil, &dr)
		} else if showBlank || (s.activeVisual != -1 && s.cfg.UseFixation) {
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
			onsetNS := state.exp.Stimuli[state.csIndex].TimestampMS * msToNS
			ctNS := sdl.TicksNS() - state.stNS
			if ctNS < onsetNS && onsetNS-ctNS <= vrrWaitNS { // If within 2ms, busy-wait
				for sdl.TicksNS()-state.stNS < onsetNS {
					// busy wait
				}
			}
		}

		trig, tidx := state.update()
		state.render()

		if trig {
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
			case StimVideo:
				stype = "VIDEO_ONSET"
			}
			state.log.Log(stim.TimestampMS, state.ctNS/msToNS, stype, label, stim.RawRow)
		}

		if !cfg.VSync && !cfg.VRR {
			sdl.Delay(1)
		}
	}

	return !state.aborted
}

