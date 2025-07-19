//go:build rp

package shared

import "machine"

// Avoid requiring packages that require specific hardware so we can run unit tests

const (
	KHz = 1_000
	MHz = 1_000_000

	NoPin = MockPin(0xff)
)

// package rp2-pio
type I2S interface {
	SetSampleFrequency(f uint32) error
	WriteMono(data []uint16) (int, error)
}

// package machine
type Pin interface {
	Configure(machine.PinConfig)
	Get() bool
	High()
	Low()
	Set(bool)
	SetInterrupt(machine.PinChange, func(machine.Pin)) error
}

// package machine
type PinConfig any

// package machine
type PWMConfig any

// package rp2-pio
type StateMachine interface {
	GetX() uint32
	IsRxFIFOEmpty() bool
	IsRxFIFOFull() bool
	IsTxFIFOEmpty() bool
	IsTxFIFOFull() bool
	RxFIFOLevel() uint32
	RxGet() uint32
	SetEnabled(bool)
	TxPut(uint32)
}

type CVCallbackFunc func(uint16, uint8) bool

type OutputCallback func(uint16, bool)

type MockPin uint8

func (m MockPin) Configure(mode machine.PinConfig) {}

func (m MockPin) Get() bool {
	return false
}

func (m MockPin) High() {}

func (m MockPin) Low() {}

func (m MockPin) Set(bool) {}

func (m MockPin) SetInterrupt(machine.PinChange, func(machine.Pin)) error {
	return nil
}
