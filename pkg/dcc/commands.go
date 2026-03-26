package dcc

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/mikesmitty/beacon-dcc/pkg/dccex"
)

// CmdThrottle handles the DCC-EX <t> command.
// Legacy format: <t REGISTER CAB SPEED DIRECTION>
// Modern format: <t CAB SPEED DIRECTION>
func (d *DCC) CmdThrottle(resp *bytes.Buffer, cmd byte, params [][]byte) error {
	var cab uint16
	var speed uint8
	var direction bool

	switch len(params) {
	case 3:
		// Modern: <t CAB SPEED DIRECTION>
		c, err := parseUint16(params[0])
		if err != nil {
			return fmt.Errorf("invalid cab: %s", params[0])
		}
		cab = c

		s, err := parseUint8(params[1])
		if err != nil {
			return fmt.Errorf("invalid speed: %s", params[1])
		}
		speed = s

		dir, err := parseUint8(params[2])
		if err != nil {
			return fmt.Errorf("invalid direction: %s", params[2])
		}
		direction = dir != 0

	case 4:
		// Legacy: <t REGISTER CAB SPEED DIRECTION> (register is ignored)
		c, err := parseUint16(params[1])
		if err != nil {
			return fmt.Errorf("invalid cab: %s", params[1])
		}
		cab = c

		s, err := parseUint8(params[2])
		if err != nil {
			return fmt.Errorf("invalid speed: %s", params[2])
		}
		speed = s

		dir, err := parseUint8(params[3])
		if err != nil {
			return fmt.Errorf("invalid direction: %s", params[3])
		}
		direction = dir != 0

	default:
		return fmt.Errorf("expected 3 or 4 parameters, got %d", len(params))
	}

	if cab == 0 {
		return fmt.Errorf("invalid cab address: 0")
	}

	// Ensure the loco exists in the state table
	d.getOrCreateLocoState(cab)

	// Build speed step byte (bit 7 = direction, bits 6-0 = speed)
	speedStep := speed & 0x7F
	if direction {
		speedStep |= 0x80
	}

	// Update state and send DCC packet
	d.SetLocoSpeedStep(cab, speedStep)
	d.setThrottle(cab, speedStep)

	// Respond with loco state broadcast
	state, _ := d.LocoState(cab)
	resp.WriteString(dccex.BroadcastLocoState(cab, state.SpeedStep, state.Functions))
	return nil
}

// CmdForgetLoco handles the DCC-EX <-> command.
// Format: <- CAB> to forget one loco, <- *> to forget all.
func (d *DCC) CmdForgetLoco(resp *bytes.Buffer, cmd byte, params [][]byte) error {
	if len(params) != 1 {
		return fmt.Errorf("expected 1 parameter, got %d", len(params))
	}

	if bytes.Equal(params[0], []byte("*")) {
		d.setThrottle(0, 1) // ESTOP all
		d.mu.RLock()
		locos := make([]uint16, 0, len(d.state))
		for loco := range d.state {
			locos = append(locos, loco)
		}
		d.mu.RUnlock()

		for _, loco := range locos {
			d.RemoveLocoState(loco)
			resp.WriteString(dccex.BroadcastForgetLoco(loco))
		}
		return nil
	}

	cab, err := parseUint16(params[0])
	if err != nil {
		return fmt.Errorf("invalid cab: %s", params[0])
	}

	loco := cab
	if _, err := d.LocoState(loco); err == nil {
		d.setThrottle(loco, 1) // Emergency stop
		d.RemoveLocoState(loco)
		d.setThrottle(loco, 1) // Emergency stop again
	}
	resp.WriteString(dccex.BroadcastForgetLoco(loco))
	return nil
}

func parseUint16(b []byte) (uint16, error) {
	v, err := strconv.ParseUint(string(b), 10, 16)
	return uint16(v), err
}

func parseUint8(b []byte) (uint8, error) {
	v, err := strconv.ParseUint(string(b), 10, 8)
	return uint8(v), err
}
