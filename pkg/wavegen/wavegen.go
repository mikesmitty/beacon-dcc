package wavegen

import (
	"github.com/mikesmitty/beacon-dcc/pkg/shared"
)

//go:generate pioasm -o go cutout.pio cutout_pio.go
//go:generate pioasm -o go wavegen.pio wavegen_pio.go

type Wavegen struct {
	signalPin shared.Pin
	brakePin  shared.Pin

	cutoutSM     shared.StateMachine
	cutoutOffset uint8
	waveSM       shared.StateMachine
	waveOffset   uint8

	railcom bool
}

func NewWavegen(pioNum int, signalPin, brakePin shared.Pin) (*Wavegen, error) {
	w := &Wavegen{
		signalPin: signalPin,
		brakePin:  brakePin,
	}

	err := w.initPIO(signalPin, brakePin)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// Enable or disable the DCC generator
func (w *Wavegen) Enable(enabled bool) {
	w.cutoutSM.SetEnabled(enabled)
	w.waveSM.SetEnabled(enabled)
	// Set the brake pin to kill power when disabled
	w.brakePin.Set(!enabled)
}
