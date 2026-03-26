//go:build !rp

package adc

import (
	"github.com/mikesmitty/beacon-dcc/pkg/motor"
	"github.com/mikesmitty/beacon-dcc/pkg/shared"
)

var _ motor.ADC = (*ADC)(nil)

type ADC struct{}

func NewADC(pin shared.Pin) *ADC {
	return &ADC{}
}

func (a *ADC) InitADC() {}

func (a *ADC) SetBaseline() {}

func (a *ADC) Get() uint16 {
	return 0
}
