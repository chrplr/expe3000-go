package engine

type StimType int

const (
	StimImage StimType = iota
	StimSound
	StimText
	StimEnd
)

type Stimulus struct {
	TimestampMS uint64
	DurationMS  uint64
	Type        StimType
	FilePath    string
	RawRow      []string
}

type Experiment struct {
	Header  []string
	Stimuli []Stimulus
}
