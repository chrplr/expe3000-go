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
	cfg.LoadCache()
	cfg.OutputFile = "" // Always start with empty results field
	cfg.SubjectID = ""  // Always start with empty subject ID field

	// Default stimuli dir if empty
	if cfg.StimuliDir == "" {
		if _, err := os.Stat("assets"); err == nil {
			cfg.StimuliDir = "assets"
		}
	}

	for {
		if engine.RunGuiSetup(cfg) {
			savedFile, err := engine.Run(cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			if savedFile != "" {
				cfg.OutputFile = savedFile
			}
		} else {
			break
		}
	}
}
