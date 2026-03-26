//go:build rp

package adc

import (
	"machine"

	"github.com/mikesmitty/beacon-dcc/pkg/motor"
	"github.com/mikesmitty/beacon-dcc/pkg/shared"
)

var _ motor.ADC = (*ADC)(nil)

type ADC struct {
	adc machine.ADC
	pin machine.Pin

	baseline uint16
}

func NewADC(pin shared.Pin) *ADC {
	a := &ADC{
		pin: pin.(machine.Pin),
	}
	return a
}

func (a *ADC) InitADC() {
	machine.InitADC()

	hw := machine.ADC{Pin: a.pin}
	hw.Configure(machine.ADCConfig{})
	a.adc = hw
}

// SetBaseline sets the baseline reading for the ADC for zero current
func (a *ADC) SetBaseline() {
	a.baseline = 0
	var sum uint32
	for range 10 {
		sum += uint32(a.Get())
	}
	a.baseline = uint16(sum / 10)
}

func (a *ADC) Get() uint16 {
	reading := a.adc.Get()
	// Make sure we don't overflow if the baseline is too high
	if a.baseline > reading {
		a.baseline = reading
	}
	return reading - a.baseline
}
