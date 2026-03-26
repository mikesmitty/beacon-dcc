//go:build !rp

package pwm

import "github.com/mikesmitty/beacon-dcc/pkg/shared"

func (s *SimplePWM) Enable(enable bool) {
}

func (s *SimplePWM) SetDuty(duty float32) {
}

func (s *SimplePWM) SetFreq(freq uint64) {
}

type pwm interface {
	Set(channel uint8, value uint64)
	SetPeriod(period uint64) error
	Enable(bool)
	Top() uint32
	Configure(config shared.PWMConfig) error
	Channel(shared.Pin) (uint8, error)
}

func NewPWM(pin shared.Pin, freq uint64, duty float32) (*SimplePWM, error) {
	return nil, nil
}
