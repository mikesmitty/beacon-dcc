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
	if len(data) > cap(p.data)-2 {
		// FIXME: Add logging around this instead? This shouldn't ever happen. There should always be room for CRC+XOR
		panic(fmt.Sprintf("data length %d exceeds max packet size %d", len(data), cap(p.data)-2))
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
	p.data = append(p.data, b...)
	if len(p.data) > cap(p.data)-2 {
		// FIXME: Add logging around this instead? This shouldn't ever happen. There should always be room for CRC+XOR
		panic(fmt.Sprintf("data length %d exceeds max packet size %d", len(p.data), cap(p.data)-2))
	}
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
