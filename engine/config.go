package engine

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/Zyko0/go-sdl3/sdl"
)

type Config struct {
	SubjectID     string    `toml:"subject_id"`
	CSVFile       string    `toml:"csv_file"`
	OutputFile    string    `toml:"output_file"`
	StimuliDir    string    `toml:"stimuli_dir"`
	StartSplash   string    `toml:"start_splash"`
	EndSplash     string    `toml:"end_splash"`
	FontFile      string    `toml:"font"`
	DLPDevice     string    `toml:"dlp"`
	FontSize      int       `toml:"font_size"`
	ScreenWidth   int       `toml:"width"`
	ScreenHeight  int       `toml:"height"`
	DisplayIndex  int       `toml:"display"`
	ScaleFactor   float32   `toml:"scale"`
	TotalDuration uint64    `toml:"total_duration"`
	UseFixation   bool      `toml:"use_fixation"`
	Fullscreen    bool      `toml:"fullscreen"`
	SkipWait      bool      `toml:"skip_wait"`
	VSync         bool      `toml:"vsync"`
	VRR           bool      `toml:"vrr"`
	BGColor       sdl.Color `toml:"-"`
	TextColor     sdl.Color `toml:"-"`
	FixationColor sdl.Color `toml:"-"`
	BGColorStr    string    `toml:"bg_color"`
	TextColorStr  string    `toml:"text_color"`
	FixColorStr   string    `toml:"fixation_color"`
}

func ParseColor(s string) sdl.Color {
	var r, g, b, a uint8
	n, _ := fmt.Sscanf(s, "%d,%d,%d,%d", &r, &g, &b, &a)
	if n < 4 && s != "" {
		a = 255
	}
	return sdl.Color{R: r, G: g, B: b, A: a}
}

const CacheFile = ".expe3000_cache"

func (cfg *Config) SaveCache() {
	cfg.BGColorStr = fmt.Sprintf("%d,%d,%d,%d", cfg.BGColor.R, cfg.BGColor.G, cfg.BGColor.B, cfg.BGColor.A)
	cfg.TextColorStr = fmt.Sprintf("%d,%d,%d,%d", cfg.TextColor.R, cfg.TextColor.G, cfg.TextColor.B, cfg.TextColor.A)
	cfg.FixColorStr = fmt.Sprintf("%d,%d,%d,%d", cfg.FixationColor.R, cfg.FixationColor.G, cfg.FixationColor.B, cfg.FixationColor.A)

	f, err := os.Create(CacheFile)
	if err != nil {
		return
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		fmt.Printf("Error saving cache: %v\n", err)
	}
}

func (cfg *Config) LoadCache() {
	data, err := os.ReadFile(CacheFile)
	if err != nil {
		return
	}

	if _, err := toml.Decode(string(data), cfg); err != nil {
		fmt.Printf("Error loading cache: %v\n", err)
		return
	}

	if cfg.BGColorStr != "" {
		cfg.BGColor = ParseColor(cfg.BGColorStr)
	}
	if cfg.TextColorStr != "" {
		cfg.TextColor = ParseColor(cfg.TextColorStr)
	}
	if cfg.FixColorStr != "" {
		cfg.FixationColor = ParseColor(cfg.FixColorStr)
	}
}

func DefaultConfig() *Config {
	return &Config{
		OutputFile:    "results.csv",
		FontSize:      50,
		ScreenWidth:   1440,
		ScreenHeight:  900,
		ScaleFactor:   1.0,
		UseFixation:   true,
		VSync:         true,
		BGColor:       sdl.Color{R: 0, G: 0, B: 0, A: 255},
		TextColor:     sdl.Color{R: 255, G: 255, B: 255, A: 255},
		FixationColor: sdl.Color{R: 255, G: 255, B: 255, A: 255},
		SkipWait:      false,
	}
}
