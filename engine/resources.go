package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Zyko0/go-sdl3/img"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

func GetDefaultFontPath() string {
	// Check local fonts directory
	entries, err := os.ReadDir("fonts")
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				ext := strings.ToLower(filepath.Ext(entry.Name()))
				if ext == ".ttf" || ext == ".ttc" {
					return filepath.Join("fonts", entry.Name())
				}
			}
		}
	}

	// System paths
	var paths []string
	switch runtime.GOOS {
	case "windows":
		paths = []string{"C:\\Windows\\Fonts\\arial.ttf"}
	case "darwin":
		paths = []string{"/System/Library/Fonts/Helvetica.ttc"}
	default:
		paths = []string{
			"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

type Resource struct {
	Textures []*sdl.Texture
	W, H     []float32
	Sounds   []SoundResource
}

type CacheEntry struct {
	Texture *sdl.Texture
	W, H    float32
	Sound   SoundResource
}

type ResourceCache struct {
	entries map[string]*CacheEntry
}

func NewResourceCache() *ResourceCache {
	return &ResourceCache{
		entries: make(map[string]*CacheEntry),
	}
}

func (c *ResourceCache) Load(renderer *sdl.Renderer, exp *Experiment, font *ttf.Font, textColor sdl.Color, stimuliDir string) ([]Resource, error) {
	resources := make([]Resource, len(exp.Stimuli))
	targetSpec := DefaultAudioSpec()

	for i, s := range exp.Stimuli {
		resources[i] = Resource{
			Textures: make([]*sdl.Texture, 0, len(s.FilePaths)),
			W:        make([]float32, 0, len(s.FilePaths)),
			H:        make([]float32, 0, len(s.FilePaths)),
			Sounds:   make([]SoundResource, 0, len(s.FilePaths)),
		}

		for _, path := range s.FilePaths {
			// Determine single asset type for caching
			assetType := s.Type
			if assetType == StimImageStream {
				assetType = StimImage
			} else if assetType == StimTextStream {
				assetType = StimText
			} else if assetType == StimSoundStream {
				assetType = StimSound
			}

			key := fmt.Sprintf("%d:%s", assetType, path)
			if entry, ok := c.entries[key]; ok {
				if entry.Texture != nil {
					resources[i].Textures = append(resources[i].Textures, entry.Texture)
					resources[i].W = append(resources[i].W, entry.W)
					resources[i].H = append(resources[i].H, entry.H)
				}
				if assetType == StimSound && entry.Sound.Data != nil {
					resources[i].Sounds = append(resources[i].Sounds, entry.Sound)
				}
				continue
			}

			entry := &CacheEntry{}
			fullPath := filepath.Join(stimuliDir, path)

			switch assetType {
			case StimImage:
				tex, err := img.LoadTexture(renderer, fullPath)
				if err != nil {
					return nil, fmt.Errorf("failed to load image: %s (%v)", fullPath, err)
				}
				entry.Texture = tex
				w, h, _ := tex.Size()
				entry.W, entry.H = w, h
			case StimSound:
				spec := &sdl.AudioSpec{}
				data, err := sdl.LoadWAV(fullPath, spec)
				if err != nil {
					return nil, fmt.Errorf("failed to load sound %s: %v", fullPath, err)
				}
				if spec.Format == targetSpec.Format && spec.Channels == targetSpec.Channels && spec.Freq == targetSpec.Freq {
					entry.Sound.Spec = *spec
					entry.Sound.Data = data
				} else {
					dstData, err := sdl.ConvertAudioSamples(spec, data, &targetSpec)
					if err != nil {
						return nil, fmt.Errorf("failed to convert sound %s to target format: %v", fullPath, err)
					}
					entry.Sound.Spec = targetSpec
					entry.Sound.Data = dstData
				}
			case StimText:
				if font != nil {
					surf, err := font.RenderTextBlended(path, textColor)
					if err != nil {
						return nil, fmt.Errorf("failed to render text '%s': %v", path, err)
					}
					if surf == nil {
						return nil, fmt.Errorf("failed to render text '%s': null surface", path)
					}
					tex, err := renderer.CreateTextureFromSurface(surf)
					if err != nil {
						surf.Destroy()
						return nil, fmt.Errorf("failed to create texture for text '%s': %v", path, err)
					}
					entry.Texture = tex
					entry.W = float32(surf.W)
					entry.H = float32(surf.H)
					surf.Destroy()
				} else {
					return nil, fmt.Errorf("cannot render text stimulus (no font loaded)")
				}
			}

			c.entries[key] = entry
			if entry.Texture != nil {
				resources[i].Textures = append(resources[i].Textures, entry.Texture)
				resources[i].W = append(resources[i].W, entry.W)
				resources[i].H = append(resources[i].H, entry.H)
			}
			if assetType == StimSound {
				resources[i].Sounds = append(resources[i].Sounds, entry.Sound)
			}
		}
	}

	return resources, nil
}

func (c *ResourceCache) Destroy() {
	for _, entry := range c.entries {
		if entry.Texture != nil {
			entry.Texture.Destroy()
		}
	}
}
