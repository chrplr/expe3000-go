package engine

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

func openURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // linux, bsd, etc.
		cmd = "xdg-open"
		args = []string{url}
	}
	return exec.Command(cmd, args...).Start()
}

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

	displayStr := strconv.Itoa(cfg.DisplayIndex)
	fontSizeStr := strconv.Itoa(cfg.FontSize)

	focusBox := -1 // 0: subject, 1: csv, 2: stimuliDir, 3: output, 4: splash, 5: font, 6: dlp, 7: display, 8: fontsize

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
		C1X           = 30
		C2X           = 450
		BoxW          = 350
		BoxH          = 30
		RowSpacing    = 60
		LabelOffset   = 25
		BrowseX       = 390
		BrowseW       = 35
		ResStartYS    = 240
		CheckSize     = 20
		CheckSpacing  = 30
		OptionsY      = 460
		StartBtnY     = 720
	)

	for {
		var e sdl.Event
		for sdl.PollEvent(&e) {
			switch e.Type {
			case sdl.EVENT_QUIT:
				return false
			case sdl.EVENT_MOUSE_BUTTON_DOWN:
				me := e.MouseButtonEvent()
				mx, my := me.X, me.Y

				focusBox = -1
				// Column 1 hit detection (0-5)
				for i := 0; i < 6; i++ {
					by := float32(40 + i*RowSpacing)
					if mx >= C1X && mx <= C1X+BoxW && my >= by && my <= by+BoxH {
						focusBox = i
						break
					}
				}
				// Column 2 hit detection (6-8)
				if focusBox == -1 {
					for i := 0; i < 3; i++ {
						by := float32(40 + i*RowSpacing)
						if mx >= C2X && mx <= C2X+BoxW && my >= by && my <= by+BoxH {
							focusBox = 6 + i
							break
						}
					}
				}

				// Help button
				if mx >= 20 && mx <= 100 && my >= StartBtnY && my <= StartBtnY+40 {
					openURL("https://chrplr.github.io/expe3000-go")
				}

				// Browse buttons in Col 1
				if mx >= BrowseX && mx <= BrowseX+BrowseW {
					for i := 1; i < 6; i++ {
						by := float32(40 + i*RowSpacing)
						if my >= by && my <= by+BoxH {
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
							case 4:
								filters := []sdl.DialogFileFilter{{Name: "Images", Pattern: "png;jpg;jpeg;bmp"}}
								cb := sdl.NewDialogFileCallback(func(fileList []string, filter int32) {
									if len(fileList) > 0 {
										cfg.StartSplash = fileList[0]
									}
								})
								sdl.ShowOpenFileDialog(cb, window, filters, "", false)
							case 5:
								filters := []sdl.DialogFileFilter{{Name: "TTF Fonts", Pattern: "ttf;ttc"}}
								cb := sdl.NewDialogFileCallback(func(fileList []string, filter int32) {
									if len(fileList) > 0 {
										cfg.FontFile = fileList[0]
									}
								})
								sdl.ShowOpenFileDialog(cb, window, filters, "", false)
							}
						}
					}
				}

				// Resolution Options (Column 2)
				for i := range resOptions {
					ry := float32(ResStartYS + i*CheckSpacing)
					if mx >= C2X && mx <= C2X+200 && my >= ry && my <= ry+CheckSize {
						selectedRes = i
					}
				}

				// Options (Column 2)
				if mx >= C2X && mx <= C2X+200 && my >= OptionsY && my <= OptionsY+CheckSize {
					cfg.UseFixation = !cfg.UseFixation
				}
				if mx >= C2X && mx <= C2X+200 && my >= OptionsY+CheckSpacing && my <= OptionsY+CheckSpacing+CheckSize {
					cfg.Fullscreen = !cfg.Fullscreen
				}
				if mx >= C2X && mx <= C2X+200 && my >= OptionsY+2*CheckSpacing && my <= OptionsY+2*CheckSpacing+CheckSize {
					cfg.SkipWait = !cfg.SkipWait
				}
				if mx >= C2X && mx <= C2X+200 && my >= OptionsY+3*CheckSpacing && my <= OptionsY+3*CheckSpacing+CheckSize {
					cfg.VRR = !cfg.VRR
				}

				// Start button
				if mx >= 350 && mx <= 450 && my >= StartBtnY && my <= StartBtnY+40 {
					if cfg.CSVFile != "" {
						cfg.ScreenWidth = resOptions[selectedRes].W
						cfg.ScreenHeight = resOptions[selectedRes].H
						if v, err := strconv.Atoi(displayStr); err == nil {
							cfg.DisplayIndex = v
						}
						if v, err := strconv.Atoi(fontSizeStr); err == nil {
							cfg.FontSize = v
						}
						cfg.SaveCache()
						return true
					}
				}

				// Quit button
				if mx >= 690 && mx <= 790 && my >= StartBtnY && my <= StartBtnY+40 {
					return false
				}
			case sdl.EVENT_TEXT_INPUT:
				ti := e.TextInputEvent()
				if focusBox != -1 {
					var target *string
					switch focusBox {
					case 0: target = &cfg.SubjectID
					case 1: target = &cfg.CSVFile
					case 2: target = &cfg.StimuliDir
					case 3: target = &cfg.OutputFile
					case 4: target = &cfg.StartSplash
					case 5: target = &cfg.FontFile
					case 6: target = &cfg.DLPDevice
					case 7: target = &displayStr
					case 8: target = &fontSizeStr
					}
					*target += ti.Text
				}
			case sdl.EVENT_KEY_DOWN:
				ke := e.KeyboardEvent()
				if focusBox != -1 {
					if ke.Key == sdl.K_BACKSPACE {
						var target *string
						switch focusBox {
						case 0: target = &cfg.SubjectID
						case 1: target = &cfg.CSVFile
						case 2: target = &cfg.StimuliDir
						case 3: target = &cfg.OutputFile
						case 4: target = &cfg.StartSplash
						case 5: target = &cfg.FontFile
						case 6: target = &cfg.DLPDevice
						case 7: target = &displayStr
						case 8: target = &fontSizeStr
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
		col1Labels := []string{"Subject ID:", "Experiment CSV:", "Stimuli Directory:", "Output Results CSV:", "Start Splash Image:", "TTF Font File:"}
		for i, label := range col1Labels {
			ly := float32(15 + i*RowSpacing)
			by := float32(40 + i*RowSpacing)
			surf, err := guiFont.RenderTextBlended(label, black)
			if err == nil && surf != nil {
				tex, err := renderer.CreateTextureFromSurface(surf)
				if err == nil {
					r := sdl.FRect{X: C1X, Y: ly, W: float32(surf.W), H: float32(surf.H)}
					renderer.RenderTexture(tex, nil, &r)
					tex.Destroy()
				}
				surf.Destroy()
			}

			renderer.SetDrawColor(255, 255, 255, 255)
			box := sdl.FRect{X: C1X, Y: by, W: BoxW, H: BoxH}
			renderer.RenderFillRect(&box)
			if focusBox == i {
				renderer.SetDrawColor(0, 120, 255, 255)
			} else {
				renderer.SetDrawColor(180, 180, 180, 255)
			}
			renderer.RenderRect(&box)

			var text string
			switch i {
			case 0: text = cfg.SubjectID
			case 1: text = cfg.CSVFile
			case 2: text = cfg.StimuliDir
			case 3: text = cfg.OutputFile
			case 4: text = cfg.StartSplash
			case 5: text = cfg.FontFile
			}
			if text != "" {
				// Only show end of path if it's too long
				displayPath := text
				if len(text) > 40 {
					displayPath = "..." + text[len(text)-37:]
				}
				surf, err := guiFont.RenderTextBlended(displayPath, black)
				if err == nil && surf != nil {
					tex, err := renderer.CreateTextureFromSurface(surf)
					if err == nil {
						r := sdl.FRect{X: C1X + 5, Y: by + 5, W: float32(surf.W), H: float32(surf.H)}
						renderer.RenderTexture(tex, nil, &r)
						tex.Destroy()
					}
					surf.Destroy()
				}
			}

			if i > 0 {
				renderer.SetDrawColor(200, 200, 200, 255)
				btn := sdl.FRect{X: BrowseX, Y: by, W: BrowseW, H: BoxH}
				renderer.RenderFillRect(&btn)
				renderer.SetDrawColor(0, 0, 0, 255)
				renderer.RenderRect(&btn)
				surf, err := guiFont.RenderTextBlended("...", black)
				if err == nil && surf != nil {
					tex, err := renderer.CreateTextureFromSurface(surf)
					if err == nil {
						r := sdl.FRect{X: BrowseX + 10, Y: by + 5, W: float32(surf.W), H: float32(surf.H)}
						renderer.RenderTexture(tex, nil, &r)
						tex.Destroy()
					}
					surf.Destroy()
				}
			}
		}

		col2Labels := []string{"DLP Device:", "Display Index:", "Font Size:"}
		for i, label := range col2Labels {
			ly := float32(15 + i*RowSpacing)
			by := float32(40 + i*RowSpacing)
			surf, err := guiFont.RenderTextBlended(label, black)
			if err == nil && surf != nil {
				tex, err := renderer.CreateTextureFromSurface(surf)
				if err == nil {
					r := sdl.FRect{X: C2X, Y: ly, W: float32(surf.W), H: float32(surf.H)}
					renderer.RenderTexture(tex, nil, &r)
					tex.Destroy()
				}
				surf.Destroy()
			}

			renderer.SetDrawColor(255, 255, 255, 255)
			box := sdl.FRect{X: C2X, Y: by, W: BoxW, H: BoxH}
			renderer.RenderFillRect(&box)
			if focusBox == 6+i {
				renderer.SetDrawColor(0, 120, 255, 255)
			} else {
				renderer.SetDrawColor(180, 180, 180, 255)
			}
			renderer.RenderRect(&box)

			var text string
			switch i {
			case 0: text = cfg.DLPDevice
			case 1: text = displayStr
			case 2: text = fontSizeStr
			}
			if text != "" {
				surf, err := guiFont.RenderTextBlended(text, black)
				if err == nil && surf != nil {
					tex, err := renderer.CreateTextureFromSurface(surf)
					if err == nil {
						r := sdl.FRect{X: C2X + 5, Y: by + 5, W: float32(surf.W), H: float32(surf.H)}
						renderer.RenderTexture(tex, nil, &r)
						tex.Destroy()
					}
					surf.Destroy()
				}
			}
		}

		// Resolution Label (Column 2)
		surfRes, err := guiFont.RenderTextBlended("Resolution:", black)
		if err == nil && surfRes != nil {
			tex, err := renderer.CreateTextureFromSurface(surfRes)
			if err == nil {
				r := sdl.FRect{X: C2X, Y: ResStartYS - 25, W: float32(surfRes.W), H: float32(surfRes.H)}
				renderer.RenderTexture(tex, nil, &r)
				tex.Destroy()
			}
			surfRes.Destroy()
		}

		// Resolution Options (Column 2)
		for i, opt := range resOptions {
			ry := float32(ResStartYS + i*CheckSpacing)
			renderer.SetDrawColor(255, 255, 255, 255)
			check := sdl.FRect{X: C2X, Y: ry, W: CheckSize, H: CheckSize}
			renderer.RenderFillRect(&check)
			renderer.SetDrawColor(0, 0, 0, 255)
			renderer.RenderRect(&check)
			if selectedRes == i {
				mark := sdl.FRect{X: C2X + 4, Y: ry + 4, W: CheckSize - 8, H: CheckSize - 8}
				renderer.SetDrawColor(0, 150, 0, 255)
				renderer.RenderFillRect(&mark)
			}
			surf, err := guiFont.RenderTextBlended(opt.Label, black)
			if err == nil && surf != nil {
				tex, err := renderer.CreateTextureFromSurface(surf)
				if err == nil {
					r := sdl.FRect{X: C2X + 30, Y: ry, W: float32(surf.W), H: float32(surf.H)}
					renderer.RenderTexture(tex, nil, &r)
					tex.Destroy()
				}
				surf.Destroy()
			}
		}

		// Options (Column 2)
		fy := float32(OptionsY)
		renderer.SetDrawColor(255, 255, 255, 255)
		fixCheck := sdl.FRect{X: C2X, Y: fy, W: CheckSize, H: CheckSize}
		renderer.RenderFillRect(&fixCheck)
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.RenderRect(&fixCheck)
		if cfg.UseFixation {
			mark := sdl.FRect{X: C2X + 4, Y: fy + 4, W: CheckSize - 8, H: CheckSize - 8}
			renderer.SetDrawColor(0, 150, 0, 255)
			renderer.RenderFillRect(&mark)
		}
		surfFix, err := guiFont.RenderTextBlended("Show fixation cross", black)
		if err == nil && surfFix != nil {
			tex, err := renderer.CreateTextureFromSurface(surfFix)
			if err == nil {
				r := sdl.FRect{X: C2X + 30, Y: fy, W: float32(surfFix.W), H: float32(surfFix.H)}
				renderer.RenderTexture(tex, nil, &r)
				tex.Destroy()
			}
			surfFix.Destroy()
		}

		fsy := float32(OptionsY + CheckSpacing)
		renderer.SetDrawColor(255, 255, 255, 255)
		fullCheck := sdl.FRect{X: C2X, Y: fsy, W: CheckSize, H: CheckSize}
		renderer.RenderFillRect(&fullCheck)
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.RenderRect(&fullCheck)
		if cfg.Fullscreen {
			mark := sdl.FRect{X: C2X + 4, Y: fsy + 4, W: CheckSize - 8, H: CheckSize - 8}
			renderer.SetDrawColor(0, 150, 0, 255)
			renderer.RenderFillRect(&mark)
		}
		surfFull, err := guiFont.RenderTextBlended("Fullscreen mode", black)
		if err == nil && surfFull != nil {
			tex, err := renderer.CreateTextureFromSurface(surfFull)
			if err == nil {
				r := sdl.FRect{X: C2X + 30, Y: fsy, W: float32(surfFull.W), H: float32(surfFull.H)}
				renderer.RenderTexture(tex, nil, &r)
				tex.Destroy()
			}
			surfFull.Destroy()
		}

		swy := float32(OptionsY + 2*CheckSpacing)
		renderer.SetDrawColor(255, 255, 255, 255)
		skipCheck := sdl.FRect{X: C2X, Y: swy, W: CheckSize, H: CheckSize}
		renderer.RenderFillRect(&skipCheck)
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.RenderRect(&skipCheck)
		if cfg.SkipWait {
			mark := sdl.FRect{X: C2X + 4, Y: swy + 4, W: CheckSize - 8, H: CheckSize - 8}
			renderer.SetDrawColor(0, 150, 0, 255)
			renderer.RenderFillRect(&mark)
		}
		surfSkip, err := guiFont.RenderTextBlended("Skip 'Press any key' screen", black)
		if err == nil && surfSkip != nil {
			tex, err := renderer.CreateTextureFromSurface(surfSkip)
			if err == nil {
				r := sdl.FRect{X: C2X + 30, Y: swy, W: float32(surfSkip.W), H: float32(surfSkip.H)}
				renderer.RenderTexture(tex, nil, &r)
				tex.Destroy()
			}
			surfSkip.Destroy()
		}

		vry := float32(OptionsY + 3*CheckSpacing)
		renderer.SetDrawColor(255, 255, 255, 255)
		vrrCheck := sdl.FRect{X: C2X, Y: vry, W: CheckSize, H: CheckSize}
		renderer.RenderFillRect(&vrrCheck)
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.RenderRect(&vrrCheck)
		if cfg.VRR {
			mark := sdl.FRect{X: C2X + 4, Y: vry + 4, W: CheckSize - 8, H: CheckSize - 8}
			renderer.SetDrawColor(0, 150, 0, 255)
			renderer.RenderFillRect(&mark)
		}
		surfVRR, err := guiFont.RenderTextBlended("Variable Refresh Rate (VRR)", black)
		if err == nil && surfVRR != nil {
			tex, err := renderer.CreateTextureFromSurface(surfVRR)
			if err == nil {
				r := sdl.FRect{X: C2X + 30, Y: vry, W: float32(surfVRR.W), H: float32(surfVRR.H)}
				renderer.RenderTexture(tex, nil, &r)
				tex.Destroy()
			}
			surfVRR.Destroy()
		}

		// Help button
		renderer.SetDrawColor(0, 120, 255, 255)
		helpBtn := sdl.FRect{X: 20, Y: StartBtnY, W: 80, H: 40}
		renderer.RenderFillRect(&helpBtn)
		white := sdl.Color{R: 255, G: 255, B: 255, A: 255}
		surfHelp, err := guiFont.RenderTextBlended("HELP", white)
		if err == nil && surfHelp != nil {
			tex, err := renderer.CreateTextureFromSurface(surfHelp)
			if err == nil {
				r := sdl.FRect{X: 20 + (80-float32(surfHelp.W))/2, Y: StartBtnY + (40-float32(surfHelp.H))/2, W: float32(surfHelp.W), H: float32(surfHelp.H)}
				renderer.RenderTexture(tex, nil, &r)
				tex.Destroy()
			}
			surfHelp.Destroy()
		}

		// Start button
		renderer.SetDrawColor(0, 150, 0, 255)
		startBtn := sdl.FRect{X: 350, Y: StartBtnY, W: 100, H: 40}
		renderer.RenderFillRect(&startBtn)
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

		// Quit button
		renderer.SetDrawColor(180, 0, 0, 255)
		quitBtn := sdl.FRect{X: 690, Y: StartBtnY, W: 100, H: 40}
		renderer.RenderFillRect(&quitBtn)
		surfQt, err := guiFont.RenderTextBlended("QUIT", white)
		if err == nil && surfQt != nil {
			tex, err := renderer.CreateTextureFromSurface(surfQt)
			if err == nil {
				r := sdl.FRect{X: 690 + (100-float32(surfQt.W))/2, Y: StartBtnY + (40-float32(surfQt.H))/2, W: float32(surfQt.W), H: float32(surfQt.H)}
				renderer.RenderTexture(tex, nil, &r)
				tex.Destroy()
			}
			surfQt.Destroy()
		}

		renderer.Present()
		sdl.Delay(10)
	}

	return false
}
