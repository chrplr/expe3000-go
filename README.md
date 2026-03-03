# expe3000 (Go Version)

**Work in progress**

A multimedia stimulus delivery system designed for experimental psychology and neuroscience tasks requiring accurate timing and low-latency audio.

This [software](http://github.com/chrplr/expe3000-go) is a port of [audiovis](https://https://chrplr.github.io/audiovis/) to Go, using the [go-sdl3](https://github.com/Zyko0/go-sdl3) bindings. A [C version of expe3000](http://https://github.com/chrplr/expe3000) also exists, with less functionnalities, which might provided better slightly better timing control in case Go is not good enough for you.
 
## Overview

Stimuli are presented according to a fixed, predefined schedule. Although keypress events are saved with a timestamp, the behavior of the program cannot be modified in real-time (e.g., immediate feedback). There is no notion of "trial." This approach is suitable for fMRI/MEG/EEG experiments with rigid stimulus presentation schedules.

## Quick Start

If you have already built the project (using `./build.sh`) or downloaded the binaries:

1. **Launch the GUI**: Run `./expe3000-gui`.
2. **Configure**: 
   - Click the **"..."** button next to **Experiment CSV** and select `experiment_new.csv`.
   - Ensure the **Stimuli Directory** points to the `assets` folder.
3. **Start**: Click the green **START** button. 
4. **Interact**: Press any key when the "Press any key to start" message appears to begin the stimulation.

Alternatively, you can run the CLI version:
```bash
./expe3000 -csv experiment.csv -stimuli-dir assets
```

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

Artifacts are named `expe3000-<version>-<os>-<arch>-binary`. Choose the one matching your system:

- **OS**: `linux`, `windows`, or `macos`.
- **Architecture**:
    - **x86_64**: For Intel or AMD 64-bit processors.
    - **arm64**: For Apple Silicon (M1/M2/M3/M4) or ARM-based Windows/Linux machines.

**How to check your architecture:**
- **Linux/macOS**: Open a terminal and run `uname -m`.
  - `x86_64` → Download the **x86_64** version.
  - `arm64` or `aarch64` → Download the **arm64** version.
- **Windows**: Open a command prompt and run `echo %PROCESSOR_ARCHITECTURE%` or check **Settings > System > About**.
  - `AMD64` → Download the **x86_64** version.
  - `ARM64` → Download the **arm64** version.

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
- `-font-size`: Font size for text stimuli (default: 50).
- `-fullscreen`: Run in fullscreen mode.
- `-no-vsync`: Disable VSync (not recommended for precise timing).
- `-no-fixation`: Disable the fixation cross.
- `-dlp`: Serial device path for DLP-IO8-G triggers (e.g., `/dev/ttyUSB0` or `COM3`).
- `-version`: Print version info and exit.

## Experiment Configuration (CSV)

The input CSV file must include at least these four columns in its header: `onset_time`, `duration`, `type`, and `stimuli`. Extra columns (like `cond`) are allowed and will be preserved in the output log.

**Example (`experiment.csv`):**
```csv
onset_time,duration,type,stimuli
1000,500,IMAGE,body01.png
2000,300,IMAGE_STREAM,face01.png:200:100~face02.png:200:100~face12.png:200:100
3000,500,TEXT,Hello !
4000,2000,BOX,Please press\nany key
7000,1,SOUND,sound02.wav
```
- **Types**: `IMAGE`, `SOUND`, `TEXT`, `BOX`, `IMAGE_STREAM`, `TEXT_STREAM`, `SOUND_STREAM`.
- **BOX**: Displays multiline text centered on the screen. Use `\n` for literal line breaks within the `stimuli` string.
- **IMAGE_STREAM**: Displays a sequence of images in rapid succession. 
    - The `stimuli` column contains filenames separated by `~`.
    - **Timing (Optional)**: Each item can use the format `filename:duration:gap`.
        - `duration`: Time in ms to show the image.
        - `gap`: Time in ms to show a blank screen (or fixation cross) after the image.
    - If timing is omitted, the value from the `duration` column is used as the frame duration with a 0ms gap.
- **TEXT_STREAM**: Displays a sequence of text strings in rapid succession. Supports the same `:duration:gap` timing format.
- **SOUND_STREAM**: Plays a sequence of sound files. Supports the same `:duration:gap` format, where `duration` is the SOA (Stimulus Onset Asynchrony).
- **Note**: For `SOUND`, use `1` (or any small value) as they play until finished, but the duration column is still required.
- **Escape**: Press **Escape** at any time to interrupt the experiment.


Note: Under Linux, you can minimize video latencies by running the cli version of expe3000  from a linux console (e.g. by pressing Ctrl-Alt-F3) and after stopping the graphics server with `systemctl stop gdm`. Thus, you will bypass x11 or wayland composers and use the Direct Rendering Manager kernel module. 


## License & Credits

Developed by [Christophe Pallier](http://www.pallier.org) <christophe@pallier.org> using [Gemini CLI](https://github.com/google/gemini-cli).

The code is distributed under the **GNU GPLv3**.

**Assets Note**: Files in the `assets/` folder are NOT public domain. Images were created by Minye Zhan and are used with permission. Do not reuse without her consent.
