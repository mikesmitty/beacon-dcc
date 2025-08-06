package wavegen

import (
	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/packet"
	"github.com/mikesmitty/beacon-dcc/pkg/shared"
)

//go:generate pioasm -o go cutout.pio cutout_pio.go
//go:generate pioasm -o go wavegen.pio wavegen_pio.go
//go:generate pioasm -o go wavegen_nocutout.pio wavegen_nocutout_pio.go
//go:generate pioasm -o go wavegen_servicemode.pio wavegen_servicemode_pio.go

type WavegenMode uint8

const (
	NormalMode WavegenMode = iota
	NoCutoutMode
	ServiceMode
)

type WavegenConfig struct {
	Mode           WavegenMode
	PioNum         int
	SignalPin      shared.Pin
	SignalPinCount uint8 // Number of adjacent signal pins to use (1 or 2)
	BrakePin       shared.Pin

	PacketReturn func(*packet.Packet) // Function to return packets to the pool after sending
}

type Wavegen struct {
	cutoutSM     shared.StateMachine
	cutoutOffset uint8
	waveSM       shared.StateMachine
	waveOffset   uint8

	WavegenConfig

	*event.EventClient
}

func NewWavegen(config WavegenConfig, cl *event.EventClient) (*Wavegen, error) {
	w := &Wavegen{
		WavegenConfig: config,
		EventClient:   cl,
	}

	err := w.initPIO(config.Mode, config.SignalPin, config.SignalPinCount, config.BrakePin)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// Enable or disable the DCC generator
func (w *Wavegen) Enable(enabled bool) {
	w.cutoutSM.SetEnabled(enabled)
	w.waveSM.SetEnabled(enabled)
	// Set the brake pin to kill power when disabled
	w.BrakePin.Set(!enabled)
}

// Generate an idle packet as needed to add space between other packets to the same decoder.
func (w *Wavegen) IdlePacket() *packet.Packet {
	p := packet.NewPacket()
	p.Fill([]byte{0xFF, 0x00}, packet.Broadcast, packet.BestEffortPriority, 0)
	return p
}

func (w *Wavegen) Update(count int) {
	// The input format is an 8 bit number containing the number of bytes in the message,
	// followed by the data bytes. For example, the standard idle packet is 0x3FF00FF
	// 3 for the length, followed by 11111111 00000000 11111111
	// The message start bit, byte terminating bits, and the packet end bit are added automatically.
	// If the FIFO is empty the statemachine will send idle packets until stopped.

	var p *packet.Packet
	var ok bool
	for range count {
		select {
		case evt := <-w.Events:
			p, ok = evt.Data.(*packet.Packet)
			if !ok {
				continue
			}

			w.send(p)
			if w.PacketReturn != nil {
				// Return the packet to the pool for reuse
				w.PacketReturn(p)
			}
		default:
			// No queued packets to process, let the PIO send idle packets
			return
		}
	}
}
