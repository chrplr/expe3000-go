package engine

import (
	"fmt"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

func RunGuiSetup(cfg *Config) bool {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_EVENTS); err != nil {
		fmt.Printf("SDL_Init Error: %v\n", err)
		return false
	}
	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		fmt.Printf("TTF_Init Error: %v\n", err)
		return false
	}
	defer ttf.Quit()

	window, renderer, err := sdl.CreateWindowAndRenderer("expe3000 Setup (Go) v2", 800, 800, 0)
	if err != nil {
		fmt.Printf("CreateWindowAndRenderer Error: %v\n", err)
		return false
	}
	defer window.Destroy()
	defer renderer.Destroy()

	fontPath := GetDefaultFontPath()
	if fontPath == "" {
		fmt.Println("Error: No default font found for GUI setup")
		return false
	}
	guiFont, err := ttf.OpenFont(fontPath, 18)
	if err != nil {
		fmt.Printf("Failed to load GUI font: %v\n", err)
		return false
	}
	defer guiFont.Close()

	setupDone := false
	focusBox := -1 // 0: subject, 1: csv, 2: stimuliDir, 3: output

	type ResOption struct {
		W, H  int
		Label string
	}
	resOptions := []ResOption{
		{800, 600, "800x600 (SVGA)"},
		{1024, 768, "1024x768 (XGA)"},
		{1366, 1024, "1366x1024 (SXGA-)"},
		{1920, 1080, "1920x1080 (FHD)"},
		{2560, 1440, "2560x1440 (QHD)"},
		{3840, 2160, "3840x2160 (4K UHD)"},
	}
	selectedRes := 3 // Default to 1080p
	for i, res := range resOptions {
		if cfg.ScreenWidth == res.W && cfg.ScreenHeight == res.H {
			selectedRes = i
			break
		}
	}

	window.StartTextInput()
	defer window.StopTextInput()

	// Layout Constants
	const (
		StartX        = 50
		BoxHeight     = 30
		ItemSpacing   = 70
		CheckSize     = 20
		CheckSpacing  = 35
		ResLabelY     = 310
		ResStartYS    = 340
		OptionsY      = 580
		StartBtnY     = 720
	)

	for !setupDone {
		var e sdl.Event
		for sdl.PollEvent(&e) {
			switch e.Type {
			case sdl.EVENT_QUIT:
				return false
			case sdl.EVENT_MOUSE_BUTTON_DOWN:
				me := e.MouseButtonEvent()
				mx, my := me.X, me.Y

				focusBox = -1
				for i := 0; i < 4; i++ {
					boxY := float32(50 + i*ItemSpacing)
					if mx >= StartX && mx <= 700 && my >= boxY && my <= boxY+BoxHeight {
						focusBox = i
						break
					}
				}

				if mx >= 710 && mx <= 780 {
					// These Y ranges are approximate for the browse buttons
					for i := 1; i < 4; i++ {
						btnY := float32(50 + i*ItemSpacing)
						if my >= btnY && my <= btnY+BoxHeight {
							switch i {
							case 1:
								filters := []sdl.DialogFileFilter{{Name: "CSV Files", Pattern: "csv"}}
								cb := sdl.NewDialogFileCallback(func(fileList []string, filter int32) {
									if len(fileList) > 0 {
										cfg.CSVFile = fileList[0]
									}
								})
								sdl.ShowOpenFileDialog(cb, window, filters, "", false)
							case 2:
								cb := sdl.NewDialogFileCallback(func(fileList []string, filter int32) {
									if len(fileList) > 0 {
										cfg.StimuliDir = fileList[0]
									}
								})
								sdl.ShowOpenFolderDialog(cb, window, "", false)
							case 3:
								cb := sdl.NewDialogFileCallback(func(fileList []string, filter int32) {
									if len(fileList) > 0 {
										cfg.OutputFile = fileList[0]
									}
								})
								sdl.ShowSaveFileDialog(cb, window, nil, "results.csv")
							}
						}
					}
				}

				for i := range resOptions {
					ry := float32(ResStartYS + i*CheckSpacing)
					if mx >= StartX && mx <= 300 && my >= ry && my <= ry+CheckSize {
						selectedRes = i
					}
				}

				if mx >= StartX && mx <= 300 && my >= OptionsY && my <= OptionsY+CheckSize {
					cfg.UseFixation = !cfg.UseFixation
				}
				if mx >= StartX && mx <= 300 && my >= OptionsY+CheckSpacing && my <= OptionsY+CheckSpacing+CheckSize {
					cfg.Fullscreen = !cfg.Fullscreen
				}

				if mx >= 350 && mx <= 450 && my >= StartBtnY && my <= StartBtnY+40 {
					if cfg.CSVFile != "" {
						cfg.ScreenWidth = resOptions[selectedRes].W
						cfg.ScreenHeight = resOptions[selectedRes].H
						cfg.SaveCache()
						setupDone = true
					}
				}
			case sdl.EVENT_TEXT_INPUT:
				ti := e.TextInputEvent()
				if focusBox != -1 {
					var target *string
					switch focusBox {
					case 0:
						target = &cfg.SubjectID
					case 1:
						target = &cfg.CSVFile
					case 2:
						target = &cfg.StimuliDir
					case 3:
						target = &cfg.OutputFile
					}
					*target += ti.Text
				}
			case sdl.EVENT_KEY_DOWN:
				ke := e.KeyboardEvent()
				if focusBox != -1 {
					if ke.Key == sdl.K_BACKSPACE {
						var target *string
						switch focusBox {
						case 0:
							target = &cfg.SubjectID
						case 1:
							target = &cfg.CSVFile
						case 2:
							target = &cfg.StimuliDir
						case 3:
							target = &cfg.OutputFile
						}
						if len(*target) > 0 {
							*target = (*target)[:len(*target)-1]
						}
					}
				}
			}
		}

		renderer.SetDrawColor(240, 240, 240, 255)
		renderer.Clear()
		black := sdl.Color{R: 0, G: 0, B: 0, A: 255}

		// Input Labels and Boxes
		labels := []string{"Subject ID:", "Experiment CSV:", "Stimuli Directory:", "Output Results CSV:"}
		for i, label := range labels {
			ly := float32(20 + i*ItemSpacing)
			surf, err := guiFont.RenderTextBlended(label, black)
			if err == nil && surf != nil {
				tex, err := renderer.CreateTextureFromSurface(surf)
				if err == nil {
					r := sdl.FRect{X: StartX, Y: ly, W: float32(surf.W), H: float32(surf.H)}
					renderer.RenderTexture(tex, nil, &r)
					tex.Destroy()
				}
				surf.Destroy()
			}

			renderer.SetDrawColor(255, 255, 255, 255)
			by := float32(50 + i*ItemSpacing)
			box := sdl.FRect{X: StartX, Y: by, W: 650, H: BoxHeight}
			renderer.RenderFillRect(&box)
			if focusBox == i {
				renderer.SetDrawColor(0, 120, 255, 255)
			} else {
				renderer.SetDrawColor(180, 180, 180, 255)
			}
			renderer.RenderRect(&box)

			var text string
			switch i {
			case 0:
				text = cfg.SubjectID
			case 1:
				text = cfg.CSVFile
			case 2:
				text = cfg.StimuliDir
			case 3:
				text = cfg.OutputFile
			}
			if text != "" {
				surf, err := guiFont.RenderTextBlended(text, black)
				if err == nil && surf != nil {
					tex, err := renderer.CreateTextureFromSurface(surf)
					if err == nil {
						r := sdl.FRect{X: StartX + 5, Y: by + 5, W: float32(surf.W), H: float32(surf.H)}
						renderer.RenderTexture(tex, nil, &r)
						tex.Destroy()
					}
					surf.Destroy()
				}
			}

			if i > 0 {
				renderer.SetDrawColor(200, 200, 200, 255)
				btn := sdl.FRect{X: 710, Y: by, W: 70, H: BoxHeight}
				renderer.RenderFillRect(&btn)
				renderer.SetDrawColor(0, 0, 0, 255)
				renderer.RenderRect(&btn)
				surf, err := guiFont.RenderTextBlended("...", black)
				if err == nil && surf != nil {
					tex, err := renderer.CreateTextureFromSurface(surf)
					if err == nil {
						r := sdl.FRect{X: 735, Y: by + 5, W: float32(surf.W), H: float32(surf.H)}
						renderer.RenderTexture(tex, nil, &r)
						tex.Destroy()
					}
					surf.Destroy()
				}
			}
		}

		// Resolution Label
		surfRes, err := guiFont.RenderTextBlended("Resolution:", black)
		if err == nil && surfRes != nil {
			tex, err := renderer.CreateTextureFromSurface(surfRes)
			if err == nil {
				r := sdl.FRect{X: StartX, Y: ResLabelY, W: float32(surfRes.W), H: float32(surfRes.H)}
				renderer.RenderTexture(tex, nil, &r)
				tex.Destroy()
			}
			surfRes.Destroy()
		}

		// Resolution Options
		for i, opt := range resOptions {
			ry := float32(ResStartYS + i*CheckSpacing)
			renderer.SetDrawColor(255, 255, 255, 255)
			check := sdl.FRect{X: StartX, Y: ry, W: CheckSize, H: CheckSize}
			renderer.RenderFillRect(&check)
			renderer.SetDrawColor(0, 0, 0, 255)
			renderer.RenderRect(&check)
			if selectedRes == i {
				mark := sdl.FRect{X: StartX + 4, Y: ry + 4, W: CheckSize - 8, H: CheckSize - 8}
				renderer.SetDrawColor(0, 150, 0, 255)
				renderer.RenderFillRect(&mark)
			}
			surf, err := guiFont.RenderTextBlended(opt.Label, black)
			if err == nil && surf != nil {
				tex, err := renderer.CreateTextureFromSurface(surf)
				if err == nil {
					r := sdl.FRect{X: StartX + 30, Y: ry, W: float32(surf.W), H: float32(surf.H)}
					renderer.RenderTexture(tex, nil, &r)
					tex.Destroy()
				}
				surf.Destroy()
			}
		}

		// Fixation checkbox
		fy := float32(OptionsY)
		renderer.SetDrawColor(255, 255, 255, 255)
		fixCheck := sdl.FRect{X: StartX, Y: fy, W: CheckSize, H: CheckSize}
		renderer.RenderFillRect(&fixCheck)
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.RenderRect(&fixCheck)
		if cfg.UseFixation {
			mark := sdl.FRect{X: StartX + 4, Y: fy + 4, W: CheckSize - 8, H: CheckSize - 8}
			renderer.SetDrawColor(0, 150, 0, 255)
			renderer.RenderFillRect(&mark)
		}
		surfFix, err := guiFont.RenderTextBlended("Show fixation cross", black)
		if err == nil && surfFix != nil {
			tex, err := renderer.CreateTextureFromSurface(surfFix)
			if err == nil {
				r := sdl.FRect{X: StartX + 30, Y: fy, W: float32(surfFix.W), H: float32(surfFix.H)}
				renderer.RenderTexture(tex, nil, &r)
				tex.Destroy()
			}
			surfFix.Destroy()
		}

		// Fullscreen checkbox
		fsy := float32(OptionsY + CheckSpacing)
		renderer.SetDrawColor(255, 255, 255, 255)
		fullCheck := sdl.FRect{X: StartX, Y: fsy, W: CheckSize, H: CheckSize}
		renderer.RenderFillRect(&fullCheck)
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.RenderRect(&fullCheck)
		if cfg.Fullscreen {
			mark := sdl.FRect{X: StartX + 4, Y: fsy + 4, W: CheckSize - 8, H: CheckSize - 8}
			renderer.SetDrawColor(0, 150, 0, 255)
			renderer.RenderFillRect(&mark)
		}
		surfFull, err := guiFont.RenderTextBlended("Fullscreen mode", black)
		if err == nil && surfFull != nil {
			tex, err := renderer.CreateTextureFromSurface(surfFull)
			if err == nil {
				r := sdl.FRect{X: StartX + 30, Y: fsy, W: float32(surfFull.W), H: float32(surfFull.H)}
				renderer.RenderTexture(tex, nil, &r)
				tex.Destroy()
			}
			surfFull.Destroy()
		}

		// Start button
		renderer.SetDrawColor(0, 150, 0, 255)
		startBtn := sdl.FRect{X: 350, Y: StartBtnY, W: 100, H: 40}
		renderer.RenderFillRect(&startBtn)
		white := sdl.Color{R: 255, G: 255, B: 255, A: 255}
		surfSt, err := guiFont.RenderTextBlended("START", white)
		if err == nil && surfSt != nil {
			tex, err := renderer.CreateTextureFromSurface(surfSt)
			if err == nil {
				r := sdl.FRect{X: 350 + (100-float32(surfSt.W))/2, Y: StartBtnY + (40-float32(surfSt.H))/2, W: float32(surfSt.W), H: float32(surfSt.H)}
				renderer.RenderTexture(tex, nil, &r)
				tex.Destroy()
			}
			surfSt.Destroy()
		}

		renderer.Present()
		sdl.Delay(10)
	}

	return true
}
