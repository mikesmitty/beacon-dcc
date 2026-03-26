//go:build !rp

package wavegen

import (
	"github.com/mikesmitty/beacon-dcc/pkg/packet"
	"github.com/mikesmitty/beacon-dcc/pkg/shared"
)

func (w *Wavegen) initPIO(mode WavegenMode, sp shared.Pin, signalPinCount uint8, bp shared.Pin) error {
	_ = sp
	_ = bp
	_ = signalPinCount
	_ = mode
	return nil
}

func (w *Wavegen) send(pkt *packet.Packet) {
	_ = pkt
}
