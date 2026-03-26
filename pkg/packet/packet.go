package packet

import (
	"fmt"
	"slices"
)

const (
	MaxPacketSize = 64 // Maximum packet size, including CRC and XOR. Should match dcc.MaxPacketSize
)

type Priority int

const (
	BestEffortPriority Priority = 0
	LowPriority        Priority = 25
	NormalPriority     Priority = 50
	HighPriority       Priority = 75
	EmergencyPriority  Priority = 100
)

const (
	Broadcast uint16 = 0xFFFF // Special pseudo-address we use for allowing broadcast packets
)

type Packet struct {
	data     []byte
	Address  uint16 // Used for tracking minimum time between packets
	Priority Priority
	Repeats  int
}

func NewPacket(size ...int) *Packet {
	if len(size) == 0 {
		size = append(size, MaxPacketSize) // Default size if not specified
	}
	return &Packet{
		data:     make([]byte, 0, size[0]),
		Priority: NormalPriority,
	}
}

func (p *Packet) Fill(data []byte, address uint16, priority Priority, repeats int) {
	max := cap(p.data) - 2
	if len(data) > max {
		data = data[:max]
	}
	p.data = p.data[:len(data)]
	copy(p.data, data)
	p.Address = address
	p.Priority = priority
	p.Repeats = repeats
}

func (p *Packet) AddByte(b byte) {
	p.AddBytes(b)
}

func (p *Packet) AddBytes(b ...byte) {
	max := cap(p.data) - 2
	if len(p.data)+len(b) > max {
		if max-len(p.data) <= 0 {
			return
		}
		b = b[:max-len(p.data)]
	}
	p.data = append(p.data, b...)
}

func (p *Packet) addChecksum() {
	if len(p.data) == 0 {
		return
	}
	checksum := byte(0)
	for _, b := range p.data {
		checksum ^= b
	}
	p.data = append(p.data, checksum)
}

func (p *Packet) IsInvalid() bool {
	if p.Len() == 0 {
		return true
	}
	if p.data[0] == 0 {
		return true
	}
	return false
}

func (p *Packet) Len() int {
	return len(p.data)
}

func (p *Packet) Bytes() []byte {
	return p.data
}

// Encode packs the packet into the PIO wavegen's uint32 word format.
//
// Word layout: each 32-bit word holds up to 4 bytes, MSB first. The first word
// reserves its high byte for the total byte count (including the XOR checksum
// that addChecksum appends), leaving room for 3 data bytes. Subsequent words
// carry 4 data bytes each.
//
// Example — idle packet [0xFF, 0x00] + checksum 0xFF = 3 bytes:
//
//	word 0: 0x03_FF_00_FF  →  [len=3][0xFF][0x00][0xFF]  (fits in one word)
//
// Example — short-address 128-step throttle [0x03, 0x3F, 0x85] + checksum 0xB9 = 4 bytes:
//
//	word 0: 0x04_03_3F_85  →  [len=4][0x03][0x3F][0x85]
//	word 1: 0xB9_00_00_00  →  [0xB9][pad][pad][pad]       (partial word, must be flushed)
//
// Because the first word only carries 3 data bytes, any packet whose total byte
// count is not a multiple of 3 will leave a partially-filled word after the loop.
// This trailing word MUST be flushed to the output — if it is dropped, the PIO
// stalls waiting for the remaining bytes declared by the length field, then
// consumes the next packet's first word as data, corrupting every subsequent
// packet on the wire.
func (p *Packet) Encode() []uint32 {
	if len(p.data) == 0 {
		return nil
	}
	p.addChecksum()
	encoded := make([]uint32, 0, (len(p.data)+4)/4)
	n := uint32(p.Len()) << 24
	i := 2
	for _, v := range p.data {
		n |= uint32(v) << (8 * i)
		i--
		if i < 0 {
			encoded = append(encoded, n)
			n = 0
			i = 3
		}
	}
	// Flush the last partial word. After the loop, i < 3 means at least one
	// byte was packed into n but the word was never appended. Dropping this
	// word causes the PIO to desync (see function comment above).
	if i < 3 {
		encoded = append(encoded, n)
	}
	return encoded
}

func (p *Packet) Reset() {
	p.data = p.data[:0]
	p.Address = 0
	p.Priority = NormalPriority
	p.Repeats = 0
}

func (p *Packet) Equal(packet *Packet) bool {
	if slices.Equal(p.data, packet.data) &&
		p.Address == packet.Address &&
		p.Priority == packet.Priority &&
		p.Repeats == packet.Repeats {
		return true
	}
	return false
}

func (p *Packet) String() string {
	return fmt.Sprintf("Packet{Address: %d, Priority: %d, Repeats: %d, Data: % x}", p.Address, p.Priority, p.Repeats, p.data)
}
