# GEMINI.md - Project Context for expe3000 (Go Version)

## Project Overview
`expe3000-go` is a high-precision multimedia stimulus delivery system designed for experimental psychology and neuroscience. It is a Go port of the original C-based `expe3000`, leveraging the SDL3 library for low-latency audio and frame-accurate visual presentation.

### Key Technologies
- **Language:** Go (v1.25+)
- **Graphics & Audio:** [SDL3](https://www.libsdl.org/) via [go-sdl3](https://github.com/Zyko0/go-sdl3) bindings.
- **Serial Communication:** [go.bug.st/serial](https://github.com/bugst/go-serial) for DLP-IO8-G trigger devices (no CGo required).
- **Architecture:**
    - `cmd/expe3000`: CLI entry point for terminal-based execution.
    - `cmd/expe3000-gui`: GUI entry point for interactive setup and execution.
    - `engine/`: Core logic including the high-precision timing loop, resource management, and CSV parsing.
    - `internal/version`: Metadata management for versioning and build info.

## Building and Running

### Build Script
The project includes a `build.sh` script that injects versioning metadata (Git tag, commit hash, and build time) into the binaries using `ldflags`.

```bash
./build.sh
```

### Manual Build
```bash
# CLI Version
go build -o expe3000 ./cmd/expe3000

# GUI Version
go build -o expe3000-gui ./cmd/expe3000-gui
```

### Testing
Unit tests are available for core components like the CSV parser and experiment validator.
```bash
go test ./engine/...
```

## Development Conventions

### High-Precision Timing Loop
The core of the system is the `RunExperiment` function in `engine/experiment.go`. It uses a predictive onset look-ahead strategy (`laMS`) and VSYNC synchronization to ensure stimuli are presented exactly when intended. 
- **Critical Section:** Garbage collection is disabled (`debug.SetGCPercent(-1)`) during the experimental loop to prevent latency spikes.
- **Event Logging:** Every onset, offset, and user response is logged with both intended and actual timestamps (in milliseconds).

### Stimulus Types
- **IMAGE / TEXT**: Standard visual stimuli.
- **SOUND**: Audio stimuli (played via a custom software mixer).
- **STREAM / TEXT_STREAM**: High-speed rapid serial visual presentation (RSVP). 
    - Multiple assets (image paths or text strings) are specified in the CSV `stimuli` column, separated by the `~` character.
    - Each frame in the stream is displayed for the duration specified in the `duration` column.

### Resource Management
- **Resource Cache:** All textures and sounds are pre-loaded into a `ResourceCache` before the experiment begins to avoid disk I/O during the critical timing loop.
- **Audio:** Uses a custom software mixer (`AudioMixer`) to ensure thread-safety and minimize startup latency.

### Configuration
- **Experiment CSV:** Schedules are defined in CSV files with columns: `onset_time`, `duration`, `type`, and `stimuli`.
- **Cache:** A `.expe3000_cache` file is used to persist the last-used settings in the GUI version.

### Platform Support
- **Linux:** Optimized for console-mode execution (via DRM) to bypass X11/Wayland overhead for maximum precision.
- **Windows/macOS:** Fully supported with automated builds via GitHub Actions.
