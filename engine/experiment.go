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
	defer w.Flush()

	w.Write([]string{"# expe3000 version: " + version.Version + " (Go version)"})
	w.Write([]string{"# Author: Christophe Pallier (christophe@pallier.org)"})
	w.Write([]string{"# GitHub: https://github.com/chrplr/expe3000"})
	w.Write([]string{"# SDL Version: " + l.SDLVersion})
	w.Write([]string{"# Platform: " + l.Platform})
	w.Write([]string{"# Hostname: " + l.Hostname})
	w.Write([]string{"# Username: " + l.Username})
	w.Write([]string{"# Video Driver: " + l.VideoDriver})
	w.Write([]string{"# Audio Driver: " + l.AudioDriver})
	w.Write([]string{"# Renderer: " + l.Renderer})
	if l.DisplayMode != "" {
		w.Write([]string{"# Display Mode: " + l.DisplayMode})
	}
	w.Write([]string{"# Logical Resolution: " + l.LogicalResolution})
	fontName := l.Font
	if fontName == "" {
		fontName = "none"
	}
	w.Write([]string{"# Font: " + fontName})
	w.Write([]string{"# Font Size: " + strconv.Itoa(l.FontSize)})
	w.Write([]string{"# Start Date: " + l.StartTime})
	w.Write([]string{"# End Date: " + l.EndTime})
	completedStr := "Completed Normally"
	if !l.Completed {
		completedStr = "Aborted (ESC or Quit)"
	}
	w.Write([]string{"# Completion Status: " + completedStr})
	w.Write([]string{"# Command Line: " + l.CommandLine})

	outputHdr := []string{"subject_id", "intended_ms", "actual_ms", "type", "label"}
	outputHdr = append(outputHdr, l.CSVHeader...)
	w.Write(outputHdr)

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

const CrossSize = 20

func drawFixationCross(renderer *sdl.Renderer, w, h int, color sdl.Color) {
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	mx, my := float32(w)/2, float32(h)/2
	renderer.RenderLine(mx-CrossSize, my, mx+CrossSize, my)
	renderer.RenderLine(mx, my-CrossSize, mx, my+CrossSize)
}

func RunExperiment(cfg *Config, exp *Experiment, resources []Resource, renderer *sdl.Renderer, mixer *AudioMixer, log *EventLog, dlp *DLPIO8G, font *ttf.Font) bool {
	// Disable Garbage Collection entirely during the critical rendering loop to avoid jitter latencies.
	prevGC := debug.SetGCPercent(-1)
	defer debug.SetGCPercent(prevGC)

	rr := float32(60.0)
	win, _ := renderer.Window()
	display := sdl.GetDisplayForWindow(win)
	mode, err := display.CurrentDisplayMode()
	if err == nil && mode.RefreshRate > 0 {
		rr = mode.RefreshRate
	}
	fdMS := uint64(1000.0 / rr)
	laMS := fdMS / 2

	stTicks := sdl.Ticks()
	cs := 0
	avi := -1
	var vet uint64
	run := true
	aborted := false

	for run {
		ct := sdl.Ticks() - stTicks

		for {
			var ev sdl.Event
			if !sdl.PollEvent(&ev) {
				break
			}
			switch ev.Type {
			case sdl.EVENT_QUIT:
				run = false
				aborted = true
			case sdl.EVENT_KEY_DOWN:
				if ev.KeyboardEvent().Key == sdl.K_ESCAPE {
					run = false
					aborted = true
				} else {
					var activeRow []string
					if avi != -1 {
						activeRow = exp.Stimuli[avi].RawRow
					} else if cs > 0 && cs-1 < len(exp.Stimuli) {
						activeRow = exp.Stimuli[cs-1].RawRow
					}
					log.Log(ct, ct, "RESPONSE", ev.KeyboardEvent().Key.KeyName(), activeRow)
				}
			}
		}

		trig := false
		tidx := -1
		if cs < len(exp.Stimuli) && (ct+laMS) >= exp.Stimuli[cs].TimestampMS {
			s := &exp.Stimuli[cs]
			if (s.Type == StimImage || s.Type == StimText || s.Type == StimStream || s.Type == StimTextStream) && len(resources[cs].Textures) > 0 {
				avi = cs
				trig = true
				tidx = cs
				if s.Type == StimStream || s.Type == StimTextStream {
					vet = ct + (s.DurationMS * uint64(len(resources[cs].Textures)))
				} else {
					vet = ct + s.DurationMS
				}
				if dlp != nil {
					if s.Type == StimImage || s.Type == StimStream {
						dlp.Set("1")
					} else {
						dlp.Set("3")
					}
				}
			} else if s.Type == StimSound && resources[cs].Sound.Data != nil {
				if mixer.Play(&resources[cs].Sound) {
					log.Log(s.TimestampMS, ct, "SOUND_ONSET", s.FilePaths[0], s.RawRow)
					if dlp != nil {
						dlp.Set("2")
						dlp.Delay(5)
						dlp.Unset("2")
					}
				}
			}
			cs++
		}

		if avi != -1 && ct >= vet {
			s := &exp.Stimuli[avi]
			intendedOff := s.TimestampMS + s.DurationMS
			if s.Type == StimStream || s.Type == StimTextStream {
				intendedOff = s.TimestampMS + (s.DurationMS * uint64(len(resources[avi].Textures)))
			}
			label := strings.Join(s.FilePaths, "~")
			stype := "IMAGE_OFFSET"
			switch s.Type {
			case StimText:
				stype = "TEXT_OFFSET"
			case StimStream:
				stype = "STREAM_OFFSET"
			case StimTextStream:
				stype = "TEXT_STREAM_OFFSET"
			}
			log.Log(intendedOff, ct, stype, label, s.RawRow)
			if dlp != nil {
				if s.Type == StimImage || s.Type == StimStream {
					dlp.Unset("1")
				} else {
					dlp.Unset("3")
				}
			}
			avi = -1
		}

		if cs >= len(exp.Stimuli) && avi == -1 && ct >= cfg.TotalDuration {
			run = false
		}

		renderer.SetDrawColor(cfg.BGColor.R, cfg.BGColor.G, cfg.BGColor.B, cfg.BGColor.A)
		renderer.Clear()
		if avi != -1 {
			r := &resources[avi]
			s := &exp.Stimuli[avi]
			
			frameIdx := 0
			if s.Type == StimStream || s.Type == StimTextStream {
				elapsed := ct - (vet - (s.DurationMS * uint64(len(r.Textures))))
				frameIdx = int(elapsed / s.DurationMS)
				if frameIdx >= len(r.Textures) {
					frameIdx = len(r.Textures) - 1
				}
				if frameIdx < 0 {
					frameIdx = 0
				}
			}

			tex := r.Textures[frameIdx]
			w := r.W[frameIdx]
			h := r.H[frameIdx]

			dr := sdl.FRect{
				X: (float32(cfg.ScreenWidth) - (w * cfg.ScaleFactor)) / 2.0,
				Y: (float32(cfg.ScreenHeight) - (h * cfg.ScaleFactor)) / 2.0,
				W: w * cfg.ScaleFactor,
				H: h * cfg.ScaleFactor,
			}
			renderer.RenderTexture(tex, nil, &dr)
		} else if cfg.UseFixation {
			drawFixationCross(renderer, cfg.ScreenWidth, cfg.ScreenHeight, cfg.FixationColor)
		}
		renderer.Present()

		if trig {
			ot := sdl.Ticks() - stTicks
			s := &exp.Stimuli[tidx]
			label := strings.Join(s.FilePaths, "~")
			stype := "IMAGE_ONSET"
			switch s.Type {
			case StimText:
				stype = "TEXT_ONSET"
			case StimStream:
				stype = "STREAM_ONSET"
			case StimTextStream:
				stype = "TEXT_STREAM_ONSET"
			}
			log.Log(s.TimestampMS, ot, stype, label, s.RawRow)
			if s.Type == StimStream || s.Type == StimTextStream {
				vet = ot + (s.DurationMS * uint64(len(resources[tidx].Textures)))
			} else {
				vet = ot + s.DurationMS
			}
		}

		if !cfg.VSync {
			sdl.Delay(1)
		}
	}

	return !aborted
}
