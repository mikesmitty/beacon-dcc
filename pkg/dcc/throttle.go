package dcc

import (
	"github.com/mikesmitty/beacon-dcc/pkg/packet"
)

func (d *DCC) SetThrottle(loco uint16, speed uint8, direction bool) {
	speedStep := speed & 0x7F
	if direction {
		speedStep |= 0x80 // Set the direction bit (bit 7)
	}

	// Schedule the throttle command
	d.setThrottle(loco, speedStep)

	// FIXME: Cleanup
	// TrackManager::setDCSignal(loco,speedStep); // in case this is a dcc track on this addr
	// // retain speed for loco reminders
	// updateLocoReminder(loco, speedStep ); // FIXME: utils.go - updates loco state I think?
}

func (d *DCC) setThrottle(loco uint16, speedStep uint8) {
	p := d.setThrottlePacket(loco, speedStep)
	d.wavegen.SendPacket(p)
}

func (d *DCC) setThrottlePacket(loco uint16, speedStep uint8) *packet.Packet {
	d.Diag("setThrottle %d %x", loco, speedStep)

	// Get a new packet with the loco address for the command
	p := d.NewPacket(loco)

	speedSteps, err := d.LocoSpeedMode(loco)
	if err != nil {
		d.Debug("error getting loco %d speed mode: %v", loco, err)
		return nil
	}

	switch speedSteps {
	case SpeedMode14:
		// 14 speed steps are not supported
	case SpeedMode28:
		speed28 := speedStep28(speedStep)
		// Construct command byte from: command speed direction
		// FIXME: Direction?
		p.AddByte(0b01000000 | speed28)
	case SpeedMode128:
		fallthrough
	default:
		p.AddBytes(cmdSetSpeed, speedStep)
	}

	// Throttle commands are high or emergency priority in case an emergency stop is needed
	p.Priority = packet.HighPriority
	if speedStep < 2 {
		p.Priority = packet.EmergencyPriority
	}
	p.Repeats = 0
	return p
}

// FIXME: Original comment:
// returns speed steps 0 to 127 (1 == emergency stop)
// or -1 on "loco not found"
func (d *DCC) getThrottleSpeed(loco uint16) (uint8, error) {
	state, err := d.LocoState(loco)
	if err != nil {
		return 0, err
	}
	// FIXME: What is this supposed to be?
	return state.SpeedStep & 0x7F, nil
}

// FIXME: Check if this is correct (it's not)
func (d *DCC) getThrottleSpeedByte(loco uint16) (uint8, error) {
	state, err := d.LocoState(loco)
	if err != nil {
		return 0, err
	}
	return state.SpeedStep, nil
}

// FIXME: Check if this is correct
func (d *DCC) getThrottleFrequency(loco uint16) (uint8, error) {
	state, err := d.LocoState(loco)
	if err != nil {
		return 0, err
	}
	// Shift out first 29 bits so we have the 3 "frequency bits" left
	res := (state.SpeedStep >> 5) & 0x07
	// d.Diag("getThrottleFrequency %d %d", loco, res) // FIXME: Cleanup?
	return uint8(res), nil
}

// FIXME: Check if this is correct
func (d *DCC) getThrottleDirection(loco uint16) (bool, error) {
	state, err := d.LocoState(loco)
	if err != nil {
		return true, err
	}
	return (state.SpeedStep & 0x80) != 0, nil
}
