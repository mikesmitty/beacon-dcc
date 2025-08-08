package motor

type PowerMode uint8

const (
	PowerModeOff PowerMode = iota
	PowerModeOn
	PowerModeAlert
	PowerModeOverload
	PowerModeNone
)

func (mode PowerMode) String() string {
	switch mode {
	case PowerModeOff:
		return "Off"
	case PowerModeOn:
		return "On"
	case PowerModeAlert:
		return "Alert"
	case PowerModeOverload:
		return "Overload"
	default:
		return "None"
	}
}
