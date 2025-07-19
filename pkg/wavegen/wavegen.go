package wavegen

import (
	"github.com/mikesmitty/beacon-dcc/pkg/shared"
)

//go:generate pioasm -o go wavegen.pio wavegen_pio.go

type Wavegen struct {
	signalPin shared.Pin
	brakePin  shared.Pin

	sm     shared.StateMachine
	offset uint8

	railcom bool
}

func NewWavegen(pioNum int, signalPin, brakePin shared.Pin) (*Wavegen, error) {
	w := &Wavegen{
		signalPin: signalPin,
		brakePin:  brakePin,
	}

	err := w.initPIO(pioNum, signalPin)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// Enable or disable the DCC generator
func (w *Wavegen) Enable(enabled bool) {
	w.sm.SetEnabled(enabled)
	// Set the brake pin to kill power when disabled
	w.brakePin.Set(!enabled)
}
