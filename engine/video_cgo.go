//go:build cgo
// +build cgo

package engine

import (
	"fmt"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/zergon321/reisen"
)

type VideoResource struct {
	Media     *reisen.Media
	Stream    *reisen.VideoStream
	Texture   *sdl.Texture
	Surface   *sdl.Surface // Reusable surface for frame conversion
	FPS       float64
	W, H      float32
	LastFrame int // Index of last decoded frame
}

func loadVideo(renderer *sdl.Renderer, fullPath string) (*VideoResource, error) {
	m, err := reisen.NewMedia(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open video %s: %v", fullPath, err)
	}

	err = m.OpenDecode()
	if err != nil {
		m.Close()
		return nil, fmt.Errorf("failed to open media for decoding %s: %v", fullPath, err)
	}

	videoStreams := m.VideoStreams()
	if len(videoStreams) == 0 {
		m.Close()
		return nil, fmt.Errorf("no video streams found in %s", fullPath)
	}
	vs := videoStreams[0]
	// Open the stream with its original width and height
	err = vs.OpenDecode(vs.Width(), vs.Height(), reisen.InterpolationFastBilinear)
	if err != nil {
		m.Close()
		return nil, fmt.Errorf("failed to open video stream in %s: %v", fullPath, err)
	}

	fps, _ := vs.FrameRate()
	if fps == 0 {
		fps = 30 // Default fallback
	}

	w, h := vs.Width(), vs.Height()
	tex, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA32, sdl.TEXTUREACCESS_STREAMING, w, h)
	if err != nil {
		vs.Close()
		m.Close()
		return nil, fmt.Errorf("failed to create streaming texture for video: %v", err)
	}

	surf, err := sdl.CreateSurface(w, h, sdl.PIXELFORMAT_RGBA32)
	if err != nil {
		tex.Destroy()
		vs.Close()
		m.Close()
		return nil, fmt.Errorf("failed to create surface for video: %v", err)
	}

	return &VideoResource{
		Media:     m,
		Stream:    vs,
		Texture:   tex,
		Surface:   surf,
		FPS:       float64(fps),
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
	if v.Stream != nil {
		v.Stream.Close()
	}
	if v.Media != nil {
		v.Media.CloseDecode()
		v.Media.Close()
	}
}

func (v *VideoResource) UpdateFrame(targetFrame int) bool {
	if targetFrame < 0 {
		targetFrame = 0
	}

	// If we need to rewind (e.g. video reused or just started)
	if targetFrame < v.LastFrame {
		v.Stream.Rewind(0)
		v.LastFrame = -1
	}

	decodedAny := false
	for v.LastFrame < targetFrame {
		frame, gotFrame, err := v.Stream.ReadVideoFrame()
		if err != nil || !gotFrame {
			break
		}
		v.LastFrame++
		decodedAny = true
		if v.LastFrame == targetFrame {
			// Update texture with new frame
			data := frame.Data()
			pixels := v.Surface.Pixels()
			copy(pixels, data)
			v.Texture.Update(nil, pixels, int32(v.Surface.Pitch))
		}
	}
	return decodedAny
}
