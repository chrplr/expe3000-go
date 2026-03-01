# expe3000 (Go Version)

**Work in progress**

A multimedia stimulus delivery system designed for experimental psychology and neuroscience tasks requiring accurate timing and low-latency audio.

This is a port of the original `expe3000` (C version) to Go, using the [go-sdl3](https://github.com/Zyko0/go-sdl3) bindings.


## Overview

Stimuli are presented according to a fixed, predefined schedule. Although keypress events are saved with a timestamp, the behavior of the program cannot be modified in real-time (e.g., immediate feedback). There is no notion of "trial." This approach is suitable for fMRI/MEG/EEG experiments with rigid stimulus presentation schedules.

## Features

- **Precise Timing:** High-resolution timing loop with VSYNC synchronization and predictive onset look-ahead.
- **Low-Latency Audio:** Uses a custom software mixer to minimize startup delay and ensure thread-safety.
- **Text Stimuli:** Support for rendering text via TTF fonts.
- **Unified Event Log:** Records stimulus onsets, offsets, and user responses in a single CSV file with a comprehensive metadata header.
- **Splashscreens:** Optional start and end screens that wait for user input.
- **Advanced Display Options:** Supports custom resolutions, logical scaling, and multiple monitors.
- **Cross-Platform:** Binaries available for Linux, Windows, and macOS (x86_64 and ARM64).
- **Serial Triggers:** Support for DLP-IO8-G devices via `go.bug.st/serial` (no CGo required).

## Prerequisites

- **Go 1.25** or later (for building from source).
- **SDL3 libraries**: 
  - **Windows**: DLLs are bundled or handled by the build system.
  - **macOS**: `brew install sdl3 sdl3_image sdl3_ttf`
  - **Linux**: Install `sdl3`, `sdl3_image`, and `sdl3_ttf` via your package manager (e.g., `apt install libsdl3-0 libsdl3-image-0 libsdl3-ttf-0`).

## Installation & Building

### Precompiled Binaries
Check the [GitHub Releases](https://github.com/chrplr/expe3000-go/releases) for automated builds for your platform.

### Building from Source
To build both the CLI and GUI versions with version metadata:
```bash
./build.sh
```

Alternatively, for a simple build:
```bash
go build -o expe3000 ./cmd/expe3000
go build -o expe3000-gui ./cmd/expe3000-gui
```

## Usage

### GUI Mode
Running the GUI version opens an **Interactive Setup Window**:
```bash
./expe3000-gui
```
- **Select Files**: Browse for Experiment CSV, Stimuli Directory, and Output file.
- **Set Resolution**: Choose from common experimental resolutions.
- **Toggle Features**: Enable/disable fixation cross and Fullscreen mode.
- **Launch**: Click **START** once the CSV path is configured.

### CLI Mode
```bash
./expe3000 -csv experiment.csv [options]
```

#### Options
- `-csv`: Path to the stimulus CSV file (required).
- `-stimuli-dir`: Directory containing image and sound assets.
- `-font`: Path to a TTF font file for text stimuli.
- `-output`: Path for the results CSV file (default: `results.csv`).
- `-width`, `-height`: Screen resolution (default: 1920x1080).
- `-fullscreen`: Run in fullscreen mode.
- `-no-vsync`: Disable VSync (not recommended for precise timing).
- `-no-fixation`: Disable the fixation cross.
- `-dlp`: Serial device path for DLP-IO8-G triggers (e.g., `/dev/ttyUSB0` or `COM3`).
- `-version`: Print version info and exit.

## Experiment Configuration (CSV)

The input CSV file must include at least these four columns in its header: `onset_time`, `duration`, `type`, and `stimuli`. Extra columns (like `cond`) are allowed and will be preserved in the output log.

**Example (`experiment.csv`):**
```csv
onset_time,duration,type,cond,stimuli
1000,500,IMAGE,Mu,Mu04.png
2000,500,IMAGE,Face,face01.png
3000,500,TEXT,Greeting,Hello !
4000,500,IMAGE,Body,body03.png
5000,1,SOUND,Animal,sound02.wav
```
- **Types**: `IMAGE`, `SOUND`, `TEXT`.
- **Note**: Use `1` (or any small value) for sounds as they play until finished, but the duration column is still required.
- **Escape**: Press **Escape** at any time to interrupt the experiment.


Note: Under Linux, you can minimize video latencies by runining the cli version of expe3000  from a linux console (e.g. by pressing Ctrl-Alt-F3) and after stopping the graphics server with `systemctl stop gdm`. Thus, you will bypass x11 or wayland composers and use the Direct Rendering Manager kernel module. 


## License & Credits

Developed by [Christophe Pallier](http://www.pallier.org) <christophe@pallier.org> using [Gemini CLI](https://github.com/google/gemini-cli).

The code is distributed under the **GNU GPLv3**.

**Assets Note**: Files in the `assets/` folder are NOT public domain. Images were created by Minye Zhan and are used with permission. Do not reuse without her consent.
