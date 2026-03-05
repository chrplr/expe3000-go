//go:build !cgo
// +build !cgo

package engine

import (
	"fmt"

	"github.com/Zyko0/go-sdl3/sdl"
)

type VideoResource struct {
	Texture   *sdl.Texture
	Surface   *sdl.Surface
	FPS       float64
	W, H      float32
	LastFrame int
}

func loadVideo(renderer *sdl.Renderer, fullPath string) (*VideoResource, error) {
	return nil, fmt.Errorf("video support is not available in this build (CGO disabled)")
}

func (v *VideoResource) Destroy() {
	if v.Texture != nil {
		v.Texture.Destroy()
	}
	if v.Surface != nil {
		v.Surface.Destroy()
	}
}

func (v *VideoResource) UpdateFrame(targetFrame int) bool {
	return false
}
