package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"expe3000/engine"
	"expe3000/internal/version"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
	"github.com/Zyko0/go-sdl3/bin/binimg"
	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/bin/binttf"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	showVersion := flag.Bool("version", false, "Print version info and exit")
	csvFile := flag.String("csv", "", "Stimulus CSV file")
	subjectID := flag.String("subject", "", "Subject ID")
	outputFile := flag.String("output", "", "Output CSV file (default: auto-generated from CSV name and timestamp)")
	stimuliDir := flag.String("stimuli-dir", "", "Directory containing stimuli")
	startSplash := flag.String("start-splash", "", "Start splash image")
	endSplash := flag.String("end-splash", "", "End splash image")
	fontFile := flag.String("font", "", "TTF font file")
	fontSize := flag.Int("font-size", 50, "Font size")
	dlpDevice := flag.String("dlp", "", "DLP-IO8-G device")
	screenW := flag.Int("width", 1440, "Screen width")
	screenH := flag.Int("height", 900, "Screen height")
	displayIdx := flag.Int("display", 0, "Display index")
	scaleFactor := flag.Float64("scale", 1.0, "Scale factor for stimuli")
	noVSync := flag.Bool("no-vsync", false, "Disable VSync")
	noFixation := flag.Bool("no-fixation", false, "Disable fixation cross")
	fullscreen := flag.Bool("fullscreen", false, "Enable fullscreen")
	vrr := flag.Bool("vrr", false, "Enable Variable Refresh Rate mode (disables VSync)")
	bgColorStr := flag.String("bg-color", "0,0,0,255", "Background color (R,G,B,A)")
	textColorStr := flag.String("text-color", "255,255,255,255", "Text color (R,G,B,A)")
	fixColorStr := flag.String("fixation-color", "255,255,255,255", "Fixation color (R,G,B,A)")
        skipWait := flag.Bool("skip-wait", false, "Skip 'Press any key to start' message")

	flag.Parse()

	if *showVersion {
		fmt.Print(version.Info())
		os.Exit(0)
	}

	defer binsdl.Load().Unload()
	defer binimg.Load().Unload()
	defer binttf.Load().Unload()

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

	// Handle Ctrl-C (SIGINT)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt, exiting...")
		// Push a quit event to the SDL queue to trigger graceful shutdown
		sdl.PushEvent(&sdl.Event{Type: sdl.EVENT_QUIT})
	}()

	cfg := engine.DefaultConfig()

	cfg.CSVFile = *csvFile
	if cfg.CSVFile == "" && flag.NArg() > 0 {
		cfg.CSVFile = flag.Arg(0)
	}
	cfg.SubjectID = *subjectID
	cfg.OutputFile = *outputFile
	cfg.StimuliDir = *stimuliDir
	cfg.StartSplash = *startSplash
	cfg.EndSplash = *endSplash
	cfg.FontFile = *fontFile
	cfg.FontSize = *fontSize
	cfg.DLPDevice = *dlpDevice
	cfg.ScreenWidth = *screenW
	cfg.ScreenHeight = *screenH
	cfg.DisplayIndex = *displayIdx
	cfg.ScaleFactor = float32(*scaleFactor)
	cfg.VSync = !*noVSync
	if *vrr {
		cfg.VSync = false
		cfg.VRR = true
	}
	cfg.UseFixation = !*noFixation
	cfg.Fullscreen = *fullscreen
	cfg.BGColor = engine.ParseColor(*bgColorStr)
	cfg.TextColor = engine.ParseColor(*textColorStr)
	cfg.FixationColor = engine.ParseColor(*fixColorStr)
	cfg.SkipWait = *skipWait

	_, err := engine.Run(cfg)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
