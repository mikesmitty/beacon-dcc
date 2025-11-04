//go:build rp

package motor

import (
	"machine"
)

var (
	ShieldEX8874 = map[string]MotorShieldProfile{
		"A": {
			PowerPin:    machine.GPIO3,
			SignalPin:   machine.GPIO12,
			BrakePin:    machine.GPIO9,
			FaultPin:    machine.GPIO4,
			SenseFactor: 1.27,
			MaxCurrent:  5000,
		},
		"B": {
			PowerPin:    machine.GPIO11,
			SignalPin:   machine.GPIO13,
			BrakePin:    machine.GPIO8,
			FaultPin:    machine.GPIO5,
			SenseFactor: 1.27,
			MaxCurrent:  5000,
		},
	}
	ArduinoMotorShieldRev3 = map[string]MotorShieldProfile{
		"A": {
			PowerPin:    machine.GPIO3,
			SignalPin:   machine.GPIO12,
			BrakePin:    machine.GPIO9,
			FaultPin:    machine.NoPin,
			MaxCurrent:  1500,
			SenseFactor: 0.488,
		},
		"B": {
			PowerPin:    machine.GPIO11,
			SignalPin:   machine.GPIO13,
			BrakePin:    machine.GPIO8,
			FaultPin:    machine.NoPin,
			MaxCurrent:  1500,
			SenseFactor: 0.488,
		},
	}
)

// FIXME: Cleanup
// #define EX8874_SHIELD F("EX8874"), \
//  new MotorDriver( 3, 12, UNUSED_PIN, 9, A0, 1.27, 5000, A4), \
//  new MotorDriver(11, 13, UNUSED_PIN, 8, A1, 1.27, 5000, A5)
// Channel          A    B
// power_pin (pwm)  3   11
// signal_pin      12   13
// signal_pin2    N/A  N/A
// brake_pin        9    8
// current_pin     A0   A1
// senseFactor   1.27 1.27
// tripMilliamps 5000 5000
// fault_pin       A4   A5
// #define STANDARD_MOTOR_SHIELD F("STANDARD_MOTOR_SHIELD"), \
// new MotorDriver( 3, 12, UNUSED_PIN, 9, A0, 0.488, 1500, UNUSED_PIN), \
// new MotorDriver(11, 13, UNUSED_PIN, 8, A1, 0.488, 1500, UNUSED_PIN)
// Channel           A     B
// power_pin (pwm)   3    11
// signal_pin       12    13
// signal_pin2     N/A   N/A
// brake_pin         9     8
// current_pin      A0    A1
// senseFactor   0.488 0.488
// tripMilliamps  1500  1500
// fault_pin       N/A   N/A
