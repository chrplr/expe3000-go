# GEMINI.md - Project Context for expe3000 (Go Version)

## Project Overview
`expe3000-go` is a high-precision multimedia stimulus delivery system designed for experimental psychology and neuroscience. It is a Go port of the original C-based `expe3000`, leveraging the SDL3 library for low-latency audio and frame-accurate visual presentation.

### Key Technologies
- **Language:** Go (v1.25+)
- **Graphics & Audio:** [SDL3](https://www.libsdl.org/) via [go-sdl3](https://github.com/Zyko0/go-sdl3) bindings.
- **Serial Communication:** [go.bug.st/serial](https://github.com/bugst/go-serial) for DLP-IO8-G trigger devices (no CGo required).
- **Configuration Parsing:** [github.com/BurntSushi/toml](https://github.com/BurntSushi/toml) for persisted settings.

## Codebase Architecture

### Entry Points
- `cmd/expe3000`: CLI entry point (`cli.go`). Handles argument parsing and terminal-based experiment execution.
- `cmd/expe3000-gui`: GUI entry point (`main.go`). Launches the interactive Fyne-based configuration window.

### Engine Components (`engine/`)
- `experiment.go`: The core high-precision timing loop. Orchestrates stimulus onset/offset, event logging, and user input handling.
- `audio.go`: Implements a custom software audio mixer. Manages audio playback buffers and ensures thread-safe, low-latency sound delivery.
- `video.go`: Handles SDL3 window management, renderer setup, and logical scaling. Support for DRM (Direct Rendering Manager) on Linux.
- `resources.go`: Implements the `ResourceCache`. Pre-loads all assets (textures, sounds, fonts) to eliminate disk I/O during the experiment.
- `stimuli.go`: Defines the stimulus data structures and provides rendering logic for images, text, and boxes.
- `csv_parser.go`: Parses stimulus schedules from CSV/TSV files, supporting complex RSVP stream formats.
- `validator.go`: Ensures the experiment CSV is logically sound (no overlapping stimuli, valid types, existing assets).
- `config.go`: Manages application settings (TOML-based persistence) and experiment parameters.
- `dlp.go`: Handles serial communication for sending triggers to DLP-IO8-G devices.
- `run.go`: Provides high-level orchestration for initializing SDL, loading resources, and running the experiment loop.

### Metadata & Versioning
- `internal/version`: Manages version strings, Git commit hashes, and build timestamps injected during compilation.

## High-Precision Timing Loop

The engine is designed for millisecond-level precision using two primary strategies:

### VSYNC Mode (Default)
- Uses a **predictive onset look-ahead** strategy. 
- The system calculates the target onset time and initiates the draw call a few milliseconds early (defined by `laMS`).
- SDL3's `SDL_RenderPresent` then waits for the monitor's next vertical refresh cycle (VSYNC) to "flip" the buffer, ensuring no screen tearing and predictable onset.

### VRR Mode (Variable Refresh Rate)
- Enabled via `--vrr`. Disables VSYNC entirely.
- Uses a **high-precision busy-wait loop** (`time.Now().UnixMilli()`) to hit the target onset millisecond exactly.
- This allows VRR-enabled (G-Sync/FreeSync) monitors to update the display immediately upon request, bypassing fixed refresh rate constraints.

### Latency Optimization
- **GC Management:** Garbage collection is disabled (`debug.SetGCPercent(-1)`) just before the experiment starts and re-enabled immediately after to prevent random latency spikes.
- **Priority:** On supported platforms (Linux console), the application can bypass Wayland/X11 to use the Direct Rendering Manager (DRM) for minimal overhead.

## Stimulus Specifications

### Types & Formatting
- **IMAGE / TEXT**: Standard single-item stimuli.
- **BOX**: Multiline text stimulus. Uses `\n` in the `stimuli` column for manual line breaks.
- **SOUND**: Audio stimuli. The `duration` column is ignored for playback (it plays to completion) but used for onset timing.
- **STREAMS (IMAGE_STREAM, TEXT_STREAM, SOUND_STREAM)**: 
    - Supports Rapid Serial Visual Presentation (RSVP).
    - Format: `item1:duration:gap~item2:duration:gap`.
    - `gap`: Time (ms) to show a blank/fixation screen between items.

### Event Logging
Each experiment generates a results file with a comprehensive metadata header including:
- Subject ID, Date/Time, Refresh Rate, and OS details.
- Full configuration parameters used during the run.

**Log Columns:**
- `onset_intended`: Requested onset time from CSV.
- `onset_actual`: True onset time as measured by the high-precision clock.
- `offset_actual`: True offset time.
- `type`, `stimuli`: Stimulus details.
- `response_key`, `response_time`: User input data.
- *Any additional columns from the input CSV are preserved.*

## Development & Testing

### Build System
The `build.sh` script is the primary way to compile the project, as it injects version metadata:
```bash
./build.sh
```

### Testing Strategy
- **Unit Tests:** `csv_parser_test.go`, `validator_test.go`, and `stream_test.go` cover the critical data ingestion and validation logic.
- **Manual Verification:** Timing accuracy should be verified using an oscilloscope or high-speed camera for sensitive experiments.

## Platform Support
- **Linux:** Optimized for TTY console execution (DRM/KMS).
- **macOS:** Fully supported (Intel & Apple Silicon). Note: Binaries may require `xattr -dr com.apple.quarantine` if downloaded.
- **Windows:** Fully supported. Includes a PowerShell installer (`install-windows.ps1`) for easy setup in `C:\Program Files`.
