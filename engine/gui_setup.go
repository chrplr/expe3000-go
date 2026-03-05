package engine

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"unicode/utf8"

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

func getTargetField(focusBox int, cfg *Config, displayStr, fontSizeStr *string) *string {
	switch focusBox {
	case 0: return &cfg.SubjectID
	case 1: return &cfg.CSVFile
	case 2: return &cfg.StimuliDir
	case 3: return &cfg.OutputFile
	case 4: return &cfg.StartSplash
	case 5: return &cfg.FontFile
	case 6: return &cfg.DLPDevice
	case 7: return displayStr
	case 8: return fontSizeStr
	default: return nil
	}
}

func renderText(renderer *sdl.Renderer, font *ttf.Font, text string, x, y float32, color sdl.Color) {
	if text == "" {
		return
	}
	surf, err := font.RenderTextBlended(text, color)
	if err != nil || surf == nil {
		return
	}
	defer surf.Destroy()
	tex, err := renderer.CreateTextureFromSurface(surf)
	if err != nil {
		return
	}
	defer tex.Destroy()
	r := sdl.FRect{X: x, Y: y, W: float32(surf.W), H: float32(surf.H)}
	renderer.RenderTexture(tex, nil, &r)
}

func renderInputBox(renderer *sdl.Renderer, font *ttf.Font, label, text string, x, y, w, h float32, isFocused bool, showBrowse bool) {
	black := sdl.Color{R: 0, G: 0, B: 0, A: 255}
	renderText(renderer, font, label, x, y-25, black)

	renderer.SetDrawColor(255, 255, 255, 255)
	box := sdl.FRect{X: x, Y: y, W: w, H: h}
	renderer.RenderFillRect(&box)
	if isFocused {
		renderer.SetDrawColor(0, 120, 255, 255)
	} else {
		renderer.SetDrawColor(180, 180, 180, 255)
	}
	renderer.RenderRect(&box)

	displayText := text
	if len(text) > 40 {
		displayText = "..." + text[len(text)-37:]
	}
	renderText(renderer, font, displayText, x+5, y+5, black)

	if showBrowse {
		renderer.SetDrawColor(200, 200, 200, 255)
		btn := sdl.FRect{X: x + w + 10, Y: y, W: 35, H: h}
		renderer.RenderFillRect(&btn)
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.RenderRect(&btn)
		renderText(renderer, font, "...", x+w+20, y+5, black)
	}
}

func renderCheckbox(renderer *sdl.Renderer, font *ttf.Font, label string, x, y float32, checked bool) {
	black := sdl.Color{R: 0, G: 0, B: 0, A: 255}
	const checkSize = 20

	renderer.SetDrawColor(255, 255, 255, 255)
	box := sdl.FRect{X: x, Y: y, W: checkSize, H: checkSize}
	renderer.RenderFillRect(&box)
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.RenderRect(&box)
	if checked {
		mark := sdl.FRect{X: x + 4, Y: y + 4, W: checkSize - 8, H: checkSize - 8}
		renderer.SetDrawColor(0, 150, 0, 255)
		renderer.RenderFillRect(&mark)
	}
	renderText(renderer, font, label, x+30, y, black)
}

func renderCenteredText(renderer *sdl.Renderer, font *ttf.Font, text string, rect sdl.FRect, color sdl.Color) {
	if text == "" {
		return
	}
	tw, th, err := font.StringSize(text)
	if err != nil {
		return
	}
	x := rect.X + (rect.W-float32(tw))/2
	y := rect.Y + (rect.H-float32(th))/2
	renderText(renderer, font, text, x, y, color)
}

func RunGuiSetup(cfg *Config) bool {
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

	focusBox := -1

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
	selectedRes := 3
	for i, res := range resOptions {
		if cfg.ScreenWidth == res.W && cfg.ScreenHeight == res.H {
			selectedRes = i
			break
		}
	}

	window.StartTextInput()
	defer window.StopTextInput()

	const (
		C1X           = 30
		C2X           = 450
		BoxW          = 350
		BoxH          = 30
		RowSpacing    = 60
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
				for i := 0; i < 6; i++ {
					by := float32(40 + i*RowSpacing)
					if mx >= C1X && mx <= C1X+BoxW && my >= by && my <= by+BoxH {
						focusBox = i
						break
					}
				}
				if focusBox == -1 {
					for i := 0; i < 3; i++ {
						by := float32(40 + i*RowSpacing)
						if mx >= C2X && mx <= C2X+BoxW && my >= by && my <= by+BoxH {
							focusBox = 6 + i
							break
						}
					}
				}

				if mx >= 20 && mx <= 100 && my >= StartBtnY && my <= StartBtnY+40 {
					openURL("https://chrplr.github.io/expe3000-go")
				}

				if mx >= BrowseX && mx <= BrowseX+BrowseW {
					for i := 1; i < 6; i++ {
						by := float32(40 + i*RowSpacing)
						if my >= by && my <= by+BoxH {
							switch i {
							case 1:
								filters := []sdl.DialogFileFilter{{Name: "Experiment Files (CSV/TSV)", Pattern: "csv;tsv"}}
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

				for i := range resOptions {
					ry := float32(ResStartYS + i*CheckSpacing)
					if mx >= C2X && mx <= C2X+200 && my >= ry && my <= ry+CheckSize {
						selectedRes = i
					}
				}

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

				if mx >= 690 && mx <= 790 && my >= StartBtnY && my <= StartBtnY+40 {
					return false
				}
			case sdl.EVENT_TEXT_INPUT:
				ti := e.TextInputEvent()
				if target := getTargetField(focusBox, cfg, &displayStr, &fontSizeStr); target != nil {
					*target += ti.Text
				}
			case sdl.EVENT_KEY_DOWN:
				ke := e.KeyboardEvent()
				if focusBox != -1 && ke.Key == sdl.K_BACKSPACE {
					if target := getTargetField(focusBox, cfg, &displayStr, &fontSizeStr); target != nil {
						if len(*target) > 0 {
							_, size := utf8.DecodeLastRuneInString(*target)
							*target = (*target)[:len(*target)-size]
						}
					}
				}
			}
		}

		renderer.SetDrawColor(240, 240, 240, 255)
		renderer.Clear()
		black := sdl.Color{R: 0, G: 0, B: 0, A: 255}

		col1Labels := []string{"Subject ID:", "Experiment CSV:", "Stimuli Directory:", "Output Results CSV:", "Start Splash Image:", "TTF Font File:"}
		for i, label := range col1Labels {
			text := ""
			switch i {
			case 0: text = cfg.SubjectID
			case 1: text = cfg.CSVFile
			case 2: text = cfg.StimuliDir
			case 3: text = cfg.OutputFile
			case 4: text = cfg.StartSplash
			case 5: text = cfg.FontFile
			}
			renderInputBox(renderer, guiFont, label, text, C1X, float32(40+i*RowSpacing), BoxW, BoxH, focusBox == i, i > 0)
		}

		col2Labels := []string{"DLP Device:", "Display Index:", "Font Size:"}
		for i, label := range col2Labels {
			text := ""
			switch i {
			case 0: text = cfg.DLPDevice
			case 1: text = displayStr
			case 2: text = fontSizeStr
			}
			renderInputBox(renderer, guiFont, label, text, C2X, float32(40+i*RowSpacing), BoxW, BoxH, focusBox == 6+i, false)
		}

		renderText(renderer, guiFont, "Resolution:", C2X, ResStartYS-25, black)
		for i, opt := range resOptions {
			renderCheckbox(renderer, guiFont, opt.Label, C2X, float32(ResStartYS+i*CheckSpacing), selectedRes == i)
		}

		renderCheckbox(renderer, guiFont, "Show fixation cross", C2X, OptionsY, cfg.UseFixation)
		renderCheckbox(renderer, guiFont, "Fullscreen mode", C2X, OptionsY+CheckSpacing, cfg.Fullscreen)
		renderCheckbox(renderer, guiFont, "Skip 'Press any key' screen", C2X, OptionsY+2*CheckSpacing, cfg.SkipWait)
		renderCheckbox(renderer, guiFont, "Variable Refresh Rate (VRR)", C2X, OptionsY+3*CheckSpacing, cfg.VRR)

		renderer.SetDrawColor(0, 120, 255, 255)
		helpBtn := sdl.FRect{X: 20, Y: StartBtnY, W: 80, H: 40}
		renderer.RenderFillRect(&helpBtn)
		white := sdl.Color{R: 255, G: 255, B: 255, A: 255}
		renderCenteredText(renderer, guiFont, "HELP", helpBtn, white)

		renderer.SetDrawColor(0, 150, 0, 255)
		startBtn := sdl.FRect{X: 350, Y: StartBtnY, W: 100, H: 40}
		renderer.RenderFillRect(&startBtn)
		renderCenteredText(renderer, guiFont, "START", startBtn, white)

		renderer.SetDrawColor(180, 0, 0, 255)
		quitBtn := sdl.FRect{X: 690, Y: StartBtnY, W: 100, H: 40}
		renderer.RenderFillRect(&quitBtn)
		renderCenteredText(renderer, guiFont, "QUIT", quitBtn, white)

		renderer.Present()
		sdl.Delay(10)
	}
}
