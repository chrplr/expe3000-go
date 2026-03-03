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
	StimEnd
)

type Stimulus struct {
	TimestampMS    uint64
	DurationMS     uint64 // Default duration for each frame or the total duration
	Type           StimType
	FilePaths      []string
	FrameDurations []uint64 // Per-frame durations (optional)
	FrameGaps      []uint64 // Per-frame gaps (optional)
	RawRow         []string
}

type Experiment struct {
	Header  []string
	Stimuli []Stimulus
}
