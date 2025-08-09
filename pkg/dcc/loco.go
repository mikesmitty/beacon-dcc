package dcc

import (
	"bytes"
	"fmt"
)

type LocoState struct {
	SpeedMode    SpeedMode
	SpeedStep    uint8
	GroupFlags   uint8
	Functions    uint32
	FuncCounter  uint16
	SpeedCounter uint16
}

// FIXME: Un-export this function
func (d *DCC) LocoState(loco uint16) (LocoState, error) {
	// Read-only lock for concurrent read access
	d.stateMutex.RLock()
	defer func() { // FIXME: Cleanup
		if r := recover(); r != nil {
			println("Recovered from panic in LocoState:")
		}
	}()
	state, ok := d.state[loco]
	d.stateMutex.RUnlock()
	if !ok {
		return LocoState{}, fmt.Errorf("state not found for loco: %d", loco)
	}
	return state, nil
}

// FIXME: Un-export this function
func (d *DCC) SetLocoState(loco uint16, state LocoState) {
	d.stateMutex.Lock()
	defer func() { // FIXME: Cleanup
		if r := recover(); r != nil {
			println("Recovered from panic in SetLocoState:")
		}
	}()
	d.state[loco] = state
	d.stateMutex.Unlock()
}

// FIXME: Un-export this function
func (d *DCC) RemoveLocoState(loco uint16) {
	d.stateMutex.Lock()
	defer func() { // FIXME: Cleanup
		if r := recover(); r != nil {
			println("Recovered from panic in RemoveLocoState:")
		}
	}()
	delete(d.state, loco)
	d.stateMutex.Unlock()
}

func (d *DCC) ForgetLoco(loco uint16) {
	d.setThrottle(loco, 1) // Emergency stop this loco if still on track
	if _, err := d.LocoState(loco); err == nil {
		d.RemoveLocoState(loco)
		d.setThrottle(loco, 1) // Emergency stop again
		// CommandDistributor::broadcastForgetLoco(loco); // FIXME: Implement
	}
}

func (d *DCC) ForgetAllLocos() {
	d.setThrottle(0, 1) // ESTOP all locos still on track
	/* FIXME: Implement
	   void DCC::forgetAllLocos() {  // removes all speed reminders
	   	  setThrottle2(0,1); // ESTOP all locos still on track
	   	  for (int i=0;i<MAX_LOCOS;i++) {
	   	    if (speedTable[i].loco) CommandDistributor::broadcastForgetLoco(speedTable[i].loco);
	   	    speedTable[i].loco=0;
	   	  }
	   	}
	*/
}

func (d *DCC) LocoSpeedMode(loco uint16) (SpeedMode, error) {
	state, err := d.LocoState(loco)
	if err != nil {
		return 0, err
	}
	return state.SpeedMode, nil
}

func (d *DCC) SetLocoSpeedMode(loco uint16, mode SpeedMode) error {
	state, err := d.LocoState(loco)
	if err != nil {
		return err
	}
	state.SpeedMode = mode
	d.SetLocoState(loco, state)
	return nil
}

func (d *DCC) LocoSpeedStep(loco uint16) (uint8, error) {
	state, err := d.LocoState(loco)
	if err != nil {
		return 0, err
	}
	return state.SpeedStep, nil
}

func (d *DCC) SetLocoSpeedStep(loco uint16, speedStep uint8) error {
	state, err := d.LocoState(loco)
	if err != nil {
		return err
	}
	state.SpeedStep = speedStep
	d.SetLocoState(loco, state)
	return nil
}

func (d *DCC) broadcastLocoState(loco uint16) {
	// FIXME: Cleanup - originally CommandDistributor::broadcastLoco
	// The original stops all locos with loco=0. Add specific logic for this instead?

	if loco == 0 {
		d.Publish("<l 0 -1 128 0>")
		return
	}

	state, err := d.LocoState(loco)
	if err != nil {
		return
	}
	// FIXME: Revisit state.SpeedStep here for 28 speed-steps
	// This byte is just the raw 128 speed step value
	d.Broadcast("<l %d 0 %d %d>", loco, state.SpeedStep, state.Functions)

	// FIXME: Implement?
	// WiThrottle::markForBroadcast(sp->loco);
}

// FIXME: Cleanup - originally DCC::displayCabList
func (d *DCC) dumpLocoState() {
	// FIXME: Revisit this
	buf := new(bytes.Buffer)
	buf.Write([]byte("<*\n"))

	d.stateMutex.RLock()
	defer func() { // FIXME: Cleanup
		if r := recover(); r != nil {
			println("Recovered from panic in dumpLocoState:")
		}
	}()
	for id, state := range d.state {
		fmt.Fprintf(buf, "cab=%d, speed=%d, functions=0x%X\n", id, state.SpeedStep, state.Functions)
	}

	fmt.Fprintf(buf, "Total=%d *>\n", len(d.state))
	d.stateMutex.RUnlock()

	d.Publish(buf)
}
