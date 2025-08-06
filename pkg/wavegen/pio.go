//go:build rp

package wavegen

import (
	"errors"
	"machine"
	"time"

	"github.com/mikesmitty/beacon-dcc/pkg/packet"
	"github.com/mikesmitty/beacon-dcc/pkg/shared"
	pio "github.com/tinygo-org/pio/rp2-pio"
)

const smFreq = 500_000

func (w *Wavegen) initPIO(mode WavegenMode, sp shared.Pin, signalPinCount uint8, bp shared.Pin) error {
	signalPin := sp.(machine.Pin)
	brakePin := bp.(machine.Pin)

	if signalPin == machine.NoPin {
		return errors.New("signal pin is not configured")
	}

	if brakePin == machine.NoPin {
		return errors.New("brake pin is not configured")
	}

	if w.waveSM != nil || w.cutoutSM != nil {
		return errors.New("wavegen already initialized")
	}

	err := w.initWavegenPIO(mode, 0, signalPin, signalPinCount)
	if err != nil {
		return err
	}

	err = w.initCutoutPIO(1, brakePin)
	if err != nil {
		return err
	}
	w.Enable(true)

	return nil
}

func (w *Wavegen) initSM(pioNum int) (pio.StateMachine, *pio.PIO, error) {
	var sm pio.StateMachine
	var err error

	switch pioNum {
	case 0:
		sm, err = pio.PIO0.ClaimStateMachine()
	case 1:
		sm, err = pio.PIO1.ClaimStateMachine()
	case 2:
		// TODO: Enable PIO2 support when available
		// sm, err = pio.PIO2.ClaimStateMachine()
		return sm, sm.PIO(), errors.New("PIO2 not yet supported")
	}
	if err != nil {
		return sm, sm.PIO(), err
	}
	Pio := sm.PIO()

	return sm, Pio, nil
}

func (w *Wavegen) initWavegenPIO(mode WavegenMode, pioNum int, signalPin machine.Pin, pinCount uint8) error {
	var instructions []uint16
	var origin int8
	var programDefaultConfig func(offset uint8) pio.StateMachineConfig

	switch mode {
	case NormalMode:
		instructions = wavegenInstructions
		origin = wavegenOrigin
		programDefaultConfig = wavegenProgramDefaultConfig
	case NoCutoutMode:
		instructions = wavegenNoCutoutInstructions
		origin = wavegenNoCutoutOrigin
		programDefaultConfig = wavegenNoCutoutProgramDefaultConfig
	case ServiceMode:
		instructions = wavegenServiceModeInstructions
		origin = wavegenServiceModeOrigin
		programDefaultConfig = wavegenServiceModeProgramDefaultConfig
	default:
		return errors.New("invalid wavegen mode")
	}

	sm, Pio, err := w.initSM(pioNum)
	if err != nil {
		return err
	}

	offset, err := Pio.AddProgram(instructions, origin)
	if err != nil {
		return err
	}

	whole, frac, err := pio.ClkDivFromFrequency(smFreq, machine.CPUFrequency())
	if err != nil {
		return err
	}

	signalPin.Configure(machine.PinConfig{Mode: Pio.PinMode()})

	cfg := programDefaultConfig(offset)
	// Disable autopush
	cfg.SetInShift(false, false, 0)
	// Enable autopull
	cfg.SetOutShift(false, true, 32)
	// Combine the TX/RX FIFO buffers to allow extra breathing room between buffer writes
	cfg.SetFIFOJoin(pio.FifoJoinTx)
	// Set set pin to the signal pins
	cfg.SetSetPins(signalPin, pinCount)
	// Enable sticky pins (set pins will remain set until cleared)
	cfg.SetOutSpecial(true, false, machine.NoPin)

	sm.Init(offset, cfg)
	sm.SetClkDiv(whole, frac)
	sm.SetPindirsConsecutive(signalPin, 1, true)

	var v uint32
	v = 0x300FF00
	// v = (1 << 24) | (0b10000001 << 16)
	sm.TxPut(v) // FIXME: Cleanup

	w.waveSM = sm
	w.waveOffset = offset

	return nil
}

func (w *Wavegen) initCutoutPIO(pioNum int, brakePin machine.Pin) error {
	sm, Pio, err := w.initSM(pioNum)
	if err != nil {
		return err
	}

	offset, err := Pio.AddProgram(cutoutInstructions, cutoutOrigin)
	if err != nil {
		return err
	}

	whole, frac, err := pio.ClkDivFromFrequency(smFreq, machine.CPUFrequency())
	if err != nil {
		return err
	}

	brakePin.Configure(machine.PinConfig{Mode: Pio.PinMode()})

	cfg := cutoutProgramDefaultConfig(offset)
	// Enable autopush
	cfg.SetInShift(false, true, 32) // FIXME: 8 bits?
	// Disable autopull
	cfg.SetOutShift(false, false, 32)
	// Combine the TX/RX FIFO buffers to allow extra breathing room between buffer writes
	cfg.SetFIFOJoin(pio.FifoJoinRx)
	// Set set pin to the brake pin
	cfg.SetSetPins(brakePin, 1)
	// Set in pin to the current flow pin
	// cfg.SetInPins(brakePin) // FIXME: Cleanup?
	// Enable sticky pins (set pins will remain set until cleared)
	cfg.SetOutSpecial(true, false, machine.NoPin)

	sm.Init(offset, cfg)
	sm.SetClkDiv(whole, frac)
	sm.SetPindirsConsecutive(brakePin, 1, true)

	w.cutoutSM = sm
	w.cutoutOffset = offset

	return nil
}

func (w *Wavegen) send(packet *packet.Packet) {
	if w.waveSM == nil || packet == nil || packet.IsInvalid() {
		// Can't send a zero-length message or to a nil state machine
		return
	}
	for _, m := range packet.Encode() {
		if !w.waveSM.IsTxFIFOFull() {
			w.waveSM.TxPut(m)
		}
		time.Sleep(1 * time.Millisecond)
	}
}
