package packet

import "sync"

type PacketPool struct {
	sync.Pool
}

func NewPacketPool(maxPacketSize int) *PacketPool {
	return &PacketPool{
		Pool: sync.Pool{
			New: func() any {
				return NewPacket(maxPacketSize)
			},
		},
	}
}

func (pp *PacketPool) NewPacket() *Packet {
	// Get a new packet from the pool
	p := pp.Get().(*Packet)
	p.Reset()
	return p
}

// Return an unused packet to the pool for reuse
func (pp *PacketPool) DiscardPacket(p *Packet) {
	if p == nil {
		return
	}
	pp.Put(p)
}
