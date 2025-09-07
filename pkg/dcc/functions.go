package dcc

import (
	"fmt"

	"github.com/mikesmitty/beacon-dcc/pkg/packet"
	"github.com/mikesmitty/beacon-dcc/pkg/topic"
)

const (
	fnGroup1 = iota + 1
	fnGroup2
	fnGroup3
	fnGroup4
	fnGroup5
)

// FIXME: Cleanup? - originally setFunctionInternal
func (d *DCC) setFunction(loco uint16, count int, b ...byte) {
	p := d.setFunctionPacket(loco, count, b...)
	d.Event.PublishTo(topic.WavegenQueue, p)
}

func (d *DCC) setFunctionPacket(loco uint16, count int, b ...byte) *packet.Packet {
	d.Event.Diag("setFunction %d % x", loco, b)

	// Get a packet with the loco address added
	p := d.NewPacket(loco)

	p.AddBytes(b...)
	p.Priority = packet.NormalPriority
	p.Repeats = count

	return p
}

// FIXME: Check if this is correct - will probably want to refactor this
func (d *DCC) setFn(loco uint16, functionNumber uint16, on bool) (bool, error) {
	// FIXME: Cleanup? Not sure why loco or fnNum are allowed to be negative
	// bool DCC::setFn( int loco, int16_t functionNumber, bool on) {
	//   if (loco<=0 ) return false;
	//   if (functionNumber < 0) return false;

	if loco == 0 {
		return false, fmt.Errorf("invalid loco number: 0")
	}

	// Get a packet with the loco address added
	p := d.NewPacket(loco)

	if functionNumber > 28 {
		// Non-reminding advanced binary bit set
		if functionNumber <= 127 {
			// Binary State Control Instruction short form
			v := uint8(functionNumber)
			if on {
				v |= 0x80
			}
			p.AddBytes(0b11011101, v)
		} else {
			// Binary State Control Instruction long form
			v := uint8(functionNumber & 0x7F)
			if on {
				v |= 0x80
			}
			// LSB byte-order
			p.AddBytes(0b11000000, v, uint8(functionNumber>>7))
		}
		p.Priority = packet.NormalPriority
		p.Repeats = 4
		d.Event.PublishTo(topic.WavegenQueue, p)
	}

	/* FIXME: Implement? Original comment:
	// We use the reminder table up to 28 for normal functions.
	// We use 29 to 31 for DC frequency as well so up to 28
	// are "real" functions and 29 to 31 are frequency bits
	// controlled by function buttons
	*/
	if functionNumber > 31 {
		return true, nil
	}

	state, err := d.LocoState(loco)
	if err != nil {
		return false, err
	}

	previous := state.Functions
	if on {
		state.Functions |= 1 << functionNumber
	} else {
		state.Functions &= ^(1 << functionNumber)
	}
	if state.Functions != previous {
		if functionNumber <= 28 {
			state.GroupFlags = d.updateGroupflags(state.GroupFlags, functionNumber)
		}
		d.SetLocoState(loco, state)
		// FIXME: Implement
		// CommandDistributor::broadcastLoco(reg)
	}

	return true, nil
}

// Flip function state (used by WiThrottle protocol)
func (d *DCC) changeFn(loco uint16, functionNumber uint16) {
	currentValue, err := d.getFn(loco, functionNumber)
	if err != nil {
		return
	}
	d.setFn(loco, functionNumber, !currentValue)
}

// Report function state (used from withrottle protocol)
// returns 0 false, 1 true or -1 for unknown
func (d *DCC) getFn(loco uint16, functionNumber uint16) (bool, error) {
	// FIXME: Cleanup? Not sure why loco or fnNum are allowed to be negative
	// if (loco<=0 || functionNumber>31)
	//   return -1;  // unknown
	// int reg = lookupSpeedTable(loco);
	// if (reg<0)
	//   return -1;
	if loco == 0 {
		return false, fmt.Errorf("invalid loco number: 0")
	} else if functionNumber > 31 {
		return false, fmt.Errorf("invalid function number: %d", functionNumber)
	}

	state, err := d.LocoState(loco)
	if err != nil {
		return false, err
	}

	return (state.Functions & (1 << functionNumber)) != 0, nil
}

// FIXME: Cleanup? - Might want to refactor this
// Set the group flag to say we have touched the particular group.
// A group will be reminded only if it has been touched.
func (d *DCC) updateGroupflags(flags byte, functionNumber uint16) byte {
	var groupMask byte
	if functionNumber <= 4 {
		groupMask = fnGroup1
	} else if functionNumber <= 8 {
		groupMask = fnGroup2
	} else if functionNumber <= 12 {
		groupMask = fnGroup3
	} else if functionNumber <= 20 {
		groupMask = fnGroup4
	} else {
		groupMask = fnGroup5
	}
	flags |= groupMask
	return flags
}

func (d *DCC) getFunctionMap(loco uint16) (uint32, error) {
	if loco == 0 {
		return 0, fmt.Errorf("invalid loco number: 0")
	}
	state, err := d.LocoState(loco)
	if err != nil {
		return 0, err
	}
	return state.Functions, nil
}
