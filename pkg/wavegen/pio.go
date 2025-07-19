//go:build rp

package wavegen

import (
	"errors"
	"machine"
	"time"

	"github.com/mikesmitty/beacon-dcc/pkg/shared"
	pio "github.com/tinygo-org/pio/rp2-pio"
)

func (w *Wavegen) initPIO(pioNum int, sp shared.Pin) error {
	var sm pio.StateMachine
	var err error
	signalPin := sp.(machine.Pin)

	switch pioNum {
	case 0:
		sm, err = pio.PIO0.ClaimStateMachine()
	case 1:
		sm, err = pio.PIO1.ClaimStateMachine()
	case 2:
		// TODO: Enable PIO2 support when available
		// sm, err = pio.PIO2.ClaimStateMachine()
		return errors.New("PIO2 not yet supported")
	}
	if err != nil {
		return err
	}
	Pio := sm.PIO()

	offset, err := Pio.AddProgram(wavegenInstructions, wavegenOrigin)
	if err != nil {
		return err
	}

	whole, frac, err := pio.ClkDivFromFrequency(smFreq, machine.CPUFrequency())
	if err != nil {
		return err
	}

	signalPin.Configure(machine.PinConfig{Mode: Pio.PinMode()})

	cfg := wavegenProgramDefaultConfig(offset)
	// Disable autopush
	cfg.SetInShift(false, false, 0)
	// Enable autopull
	cfg.SetOutShift(false, true, 32)
	// Combine the TX/RX FIFO buffers to allow extra breathing room between buffer writes
	cfg.SetFIFOJoin(pio.FifoJoinTx)
	// Set set pin to the signal pin
	cfg.SetSetPins(signalPin, 1)
	// Enable sticky pins (set pins will remain set until cleared)
	cfg.SetOutSpecial(true, false, machine.NoPin)

	sm.Init(offset, cfg)
	sm.SetClkDiv(whole, frac)
	sm.SetPindirsConsecutive(signalPin, 1, true)

	var v uint32
	v = 0x300FF00
	// v = (1 << 24) | (0b10000001 << 16)
	sm.TxPut(v) // FIXME: Cleanup

	w.sm = sm
	w.offset = offset
	w.Enable(true)

	return nil
}

func (w *Wavegen) send(msg ...uint32) {
	if w.sm == nil || len(msg) == 0 || msg[0] == 0 {
		// Can't send a zero-length message or to a nil state machine
		return
	}
	for _, m := range msg {
		for w.sm.IsTxFIFOFull() {
			time.Sleep(1 * time.Millisecond)
		}
		w.sm.TxPut(m)
	}
}
