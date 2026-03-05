package engine

import (
	"fmt"
	"io"
	"os"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/funatsufumiya/go-gv-video/gvvideo"
)

type VideoResource struct {
	Video     *gvvideo.GVVideo
	File      *os.File
	Texture   *sdl.Texture
	Surface   *sdl.Surface // Reusable surface for frame conversion
	FPS       float64
	W, H      float32
	LastFrame int // Index of last decoded frame
}

func loadVideo(renderer *sdl.Renderer, fullPath string) (*VideoResource, error) {
	// gvvideo.LoadGVVideo opens the file and reads headers/index
	v, err := gvvideo.LoadGVVideo(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load GV video %s: %v", fullPath, err)
	}

	w, h := int32(v.Header.Width), int32(v.Header.Height)

	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA32, sdl.TEXTUREACCESS_STREAMING, int(w), int(h))
	if err != nil {
		if closer, ok := v.Reader.(io.Closer); ok {
			closer.Close()
		}
		return nil, fmt.Errorf("failed to create streaming texture for video: %v", err)
	}

	surf, err := sdl.CreateSurface(int(w), int(h), sdl.PIXELFORMAT_RGBA32)
	if err != nil {
		tex.Destroy()
		if closer, ok := v.Reader.(io.Closer); ok {
			closer.Close()
		}
		return nil, fmt.Errorf("failed to create surface for video: %v", err)
	}

	return &VideoResource{
		Video:     v,
		Texture:   tex,
		Surface:   surf,
		FPS:       float64(v.Header.FPS),
		W:         float32(w),
		H:         float32(h),
		LastFrame: -1,
	}, nil
}

func (v *VideoResource) Destroy() {
	if v.Texture != nil {
		v.Texture.Destroy()
	}
	if v.Surface != nil {
		v.Surface.Destroy()
	}
	if v.Video != nil && v.Video.Reader != nil {
		if closer, ok := v.Video.Reader.(io.Closer); ok {
			closer.Close()
		}
	}
}

func (v *VideoResource) UpdateFrame(targetFrame int) bool {
	if targetFrame < 0 {
		targetFrame = 0
	}
	if targetFrame >= int(v.Video.Header.FrameCount) {
		targetFrame = int(v.Video.Header.FrameCount) - 1
	}

	if targetFrame == v.LastFrame {
		return false
	}

	// ReadFrame returns []uint8 (RGBA)
	rgba, err := v.Video.ReadFrame(uint32(targetFrame))
	if err != nil {
		return false
	}

	pixels := v.Surface.Pixels()
	copy(pixels, rgba)
	v.Texture.Update(nil, pixels, int32(v.Surface.Pitch))

	v.LastFrame = targetFrame
	return true
}
