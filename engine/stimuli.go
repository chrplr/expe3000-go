package engine

type StimType int

const (
	StimImage StimType = iota
	StimSound
	StimText
	StimImageStream
	StimTextStream
	StimSoundStream
	StimBox
	StimVideo
	StimEnd
)

func (t StimType) String() string {
	switch t {
	case StimImage:
		return "IMAGE"
	case StimSound:
		return "SOUND"
	case StimText:
		return "TEXT"
	case StimImageStream:
		return "IMAGE_STREAM"
	case StimTextStream:
		return "TEXT_STREAM"
	case StimSoundStream:
		return "SOUND_STREAM"
	case StimBox:
		return "BOX"
	case StimVideo:
		return "VIDEO"
	case StimEnd:
		return "END"
	default:
		return "UNKNOWN"
	}
}

type Stimulus struct {
	TimestampMS    uint64
	DurationMS     uint64 // Default duration for each frame or the total duration
	Type           StimType
	FilePaths      []string
	FrameDurations []uint64 // Per-frame durations (optional)
	FrameGaps      []uint64 // Per-frame gaps (optional)
	RawRow         []string
}

func (s *Stimulus) TotalDuration() uint64 {
	if s.Type == StimImageStream || s.Type == StimTextStream || s.Type == StimSoundStream {
		total := uint64(0)
		for i := 0; i < len(s.FrameDurations); i++ {
			total += s.FrameDurations[i]
			if i < len(s.FrameGaps) {
				total += s.FrameGaps[i]
			}
		}
		return total
	}
	return s.DurationMS
}

type Experiment struct {
	Header  []string
	Stimuli []Stimulus
}
