package engine

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

func Run(cfg *Config) {
	if cfg.CSVFile == "" {
		fmt.Println("Error: CSV file is required.")
		os.Exit(1)
	}

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO | sdl.INIT_EVENTS); err != nil {
		fmt.Printf("SDL_Init Error: %v\n", err)
		os.Exit(1)
	}
	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		fmt.Printf("TTF_Init Error: %v\n", err)
		os.Exit(1)
	}
	defer ttf.Quit()

	windowFlags := sdl.WINDOW_RESIZABLE
	if cfg.Fullscreen {
		windowFlags |= sdl.WINDOW_FULLSCREEN
	}

	window, renderer, err := sdl.CreateWindowAndRenderer("expe3000 (Go)", cfg.ScreenWidth, cfg.ScreenHeight, windowFlags)
	if err != nil {
		fmt.Printf("CreateWindowAndRenderer Error: %v\n", err)
		os.Exit(1)
	}
	defer window.Destroy()
	defer renderer.Destroy()

	if cfg.VSync {
		renderer.SetVSync(1)
	} else {
		renderer.SetVSync(0)
	}

	var font *ttf.Font
	if cfg.FontFile != "" {
		font, err = ttf.OpenFont(cfg.FontFile, float32(cfg.FontSize))
		if err != nil {
			fmt.Printf("Failed to load font: %s (%v)\n", cfg.FontFile, err)
			cfg.FontFile = "" // Clear it if it failed to load
		}
	}

	// If no font loaded yet (either none specified or loading failed), try default
	if font == nil {
		fontPath := GetDefaultFontPath()
		if fontPath != "" {
			font, err = ttf.OpenFont(fontPath, float32(cfg.FontSize))
			if err != nil {
				fmt.Printf("Failed to load default font: %s (%v)\n", fontPath, err)
			} else {
				cfg.FontFile = fontPath
			}
		}
	}
	defer func() {
		if font != nil {
			font.Close()
		}
	}()

	exp, err := LoadExperiment(cfg.CSVFile)
	if err != nil {
		fmt.Printf("Failed to load experiment: %v\n", err)
		os.Exit(1)
	}

	validationErrs := ValidateExperiment(exp, cfg.StimuliDir)
	if len(validationErrs) > 0 {
		fmt.Println("Experiment configuration contains errors:")
		for _, vErr := range validationErrs {
			fmt.Printf("- %v\n", vErr)
		}
		os.Exit(1)
	}

	if len(exp.Stimuli) > 0 {
		lastStim := exp.Stimuli[len(exp.Stimuli)-1]
		cfg.TotalDuration = lastStim.TimestampMS + lastStim.DurationMS + 500
	}

	cache := NewResourceCache()
	defer cache.Destroy()

	resources, err := cache.Load(renderer, exp, font, cfg.TextColor, cfg.StimuliDir)
	if err != nil {
		fmt.Printf("Failed to load resources: %v\n", err)
		os.Exit(1)
	}

	mixer := NewAudioMixer()
	spec := DefaultAudioSpec()
	cb := sdl.NewAudioStreamCallback(mixer.Callback)
	stream := sdl.AUDIO_DEVICE_DEFAULT_PLAYBACK.OpenAudioDeviceStream(&spec, cb)
	if stream == nil {
		fmt.Printf("Failed to open audio stream\n")
		os.Exit(1)
	}
	defer stream.Destroy()
	stream.ResumeDevice()

	var dlp *DLPIO8G
	if cfg.DLPDevice != "" {
		dlp, err = NewDLPIO8G(cfg.DLPDevice, 9600)
		if err != nil {
			fmt.Printf("Failed to initialize DLP device: %v\n", err)
		} else {
			defer dlp.Close()
		}
	}

	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("LOGNAME")
	}
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	if username == "" {
		username = "unknown"
	}

	sdlVer := sdl.GetVersion()
	sdlVersionStr := fmt.Sprintf("%d.%d.%d", sdlVer/1000000, (sdlVer/1000)%1000, sdlVer%1000)

	displayModeStr := ""
	if window != nil {
		display := sdl.GetDisplayForWindow(window)
		if dm, err := display.CurrentDisplayMode(); err == nil {
			displayModeStr = fmt.Sprintf("%dx%d @ %.2fHz (Physical)", dm.W, dm.H, dm.RefreshRate)
		}
	}

	rendererName, _ := renderer.Name()

	log := &EventLog{
		SubjectID:         cfg.SubjectID,
		CSVHeader:         exp.Header,
		Entries:           make([]EventLogEntry, 0, len(exp.Stimuli)*4+100),
		SDLVersion:        sdlVersionStr,
		Platform:          runtime.GOOS,
		Hostname:          hostname,
		Username:          username,
		VideoDriver:       sdl.GetCurrentVideoDriver(),
		AudioDriver:       sdl.GetCurrentAudioDriver(),
		Renderer:          rendererName,
		DisplayMode:       displayModeStr,
		LogicalResolution: fmt.Sprintf("%dx%d", cfg.ScreenWidth, cfg.ScreenHeight),
		Font:              cfg.FontFile,
		FontSize:          cfg.FontSize,
		CommandLine:       strings.Join(os.Args, " "),
	}

	if !DisplaySplash(renderer, cfg.StartSplash, cfg.ScreenWidth, cfg.ScreenHeight, cfg.ScaleFactor, cfg.BGColor) {
		return
	}

	log.StartTime = time.Now().Format("2006-01-02 15:04:05.000")

	success := RunExperiment(cfg, exp, resources, renderer, mixer, log, dlp, font)

	if success {
		log.Completed = true
		DisplaySplash(renderer, cfg.EndSplash, cfg.ScreenWidth, cfg.ScreenHeight, cfg.ScaleFactor, cfg.BGColor)
	}

	log.EndTime = time.Now().Format("2006-01-02 15:04:05.000")

	timestamp := time.Now().Format("20060102-150405")
	outputName := strings.Replace(cfg.OutputFile, ".csv", "_"+timestamp+".csv", 1)
	if err := log.Save(outputName); err != nil {
		fmt.Printf("Failed to save event log: %v\n", err)
	} else {
		fmt.Printf("\nResults saved to %s\n", outputName)
	}
}
