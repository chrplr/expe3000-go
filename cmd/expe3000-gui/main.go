package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

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
	flag.Parse()

	if *showVersion {
		fmt.Print(version.Info())
		os.Exit(0)
	}

	defer binsdl.Load().Unload()
	defer binimg.Load().Unload()
	defer binttf.Load().Unload()

	cfg := engine.DefaultConfig()
	cfg.LoadCache()

	// Default stimuli dir if empty
	if cfg.StimuliDir == "" {
		if _, err := os.Stat("assets"); err == nil {
			cfg.StimuliDir = "assets"
		}
	}

	if engine.RunGuiSetup(cfg) {
		engine.Run(cfg)
	}
}
