package pwm

import "math"

type SimplePWM struct {
	channel uint8
	pwm     pwm
	slice   uint8
	top     float32
}

const MaxFreq = math.MaxUint64
