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

func (a *ADC) Get() uint16 {
	return a.adc.Get()
}
