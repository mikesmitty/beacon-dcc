package track

type TrackMode uint8

const (
	TrackModeMain TrackMode = 1 << iota
	TrackModeProg
	TrackModeDC
)

func (tm TrackMode) String() string {
	switch tm {
	case TrackModeMain:
		return "main"
	case TrackModeProg:
		return "prog"
	case TrackModeDC:
		return "dc"
	default:
		return "unknown"
	}
}

func (tm TrackMode) Matches(other TrackMode) bool {
	return tm&other != 0
}

func (tm TrackMode) IsValid() bool {
	return tm == TrackModeMain || tm == TrackModeProg || tm == TrackModeDC
}

func (tm TrackMode) IsMain() bool {
	return tm == TrackModeMain
}

func (tm TrackMode) IsProg() bool {
	return tm&TrackModeProg != 0
}

func (tm TrackMode) IsJoined() bool {
	return tm.IsProg() && tm&TrackModeMain != 0
}

func (tm TrackMode) IsDC() bool {
	return tm == TrackModeDC
}
