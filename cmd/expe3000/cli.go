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
	outputFile := flag.String("output", "results.csv", "Output CSV file")
	stimuliDir := flag.String("stimuli-dir", "", "Directory containing stimuli")
	startSplash := flag.String("start-splash", "", "Start splash image")
	endSplash := flag.String("end-splash", "", "End splash image")
	fontFile := flag.String("font", "", "TTF font file")
	fontSize := flag.Int("font-size", 50, "Font size")
	dlpDevice := flag.String("dlp", "", "DLP-IO8-G device")
	screenW := flag.Int("width", 1920, "Screen width")
	screenH := flag.Int("height", 1080, "Screen height")
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

	// Handle Ctrl-C (SIGINT)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt, exiting...")
		os.Exit(0)
	}()

	defer binsdl.Load().Unload()
	defer binimg.Load().Unload()
	defer binttf.Load().Unload()

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

	engine.Run(cfg)
}
