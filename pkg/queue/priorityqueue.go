package queue

import (
	"container/heap"

	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/packet"
)

type PriorityQueue struct {
	items []*packet.Packet

	*event.EventClient
}

var _ heap.Interface = (*PriorityQueue)(nil)

func NewPriorityQueue(size int, eventClient *event.EventClient) PriorityQueue {
	q := PriorityQueue{
		items:       make([]*packet.Packet, 0, size),
		EventClient: eventClient,
	}
	return q
}

func (pq PriorityQueue) Len() int {
	return len(pq.items)
}

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest priority, not lowest, so we use greater than here.
	return pq.items[i].Priority > pq.items[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

func (pq *PriorityQueue) Push(x any) {
	item := x.(*packet.Packet)
	pq.items = append(pq.items, item)
}

func (pq *PriorityQueue) Pop() any {
	old := pq.items
	n := len(old)

	// If the packet still needs to be repeated, decrement and leave it in the queue
	packet := old[n-1]
	// FIXME: Does this make the most sense vs. just returning the packet and handling repeats in the caller?
	if packet.Repeats > 0 {
		packet.Repeats--
		return packet
	}

	// Remove the packet from the queue
	old[n-1] = nil // Don't prevent the GC from reclaiming this item
	pq.items = old[0 : n-1]
	return packet
}

func (pq *PriorityQueue) PushPacket(packet *packet.Packet) {
	heap.Push(pq, packet)
}

func (pq *PriorityQueue) PopPacket() *packet.Packet {
	if pq.Len() == 0 {
		// No packets to pop
		return nil
	}
	return heap.Pop(pq).(*packet.Packet)
}

func (pq *PriorityQueue) Peek() *packet.Packet {
	if pq.Len() == 0 {
		// No packets to peek
		return nil
	}
	return pq.items[0]
}

func (pq *PriorityQueue) IsEmpty() bool {
	return pq.Len() == 0
}

func (pq *PriorityQueue) Clear() {
	pq.items = make([]*packet.Packet, 0)
}

func (pq *PriorityQueue) Contains(packet *packet.Packet) bool {
	for _, p := range pq.items {
		if p.Equal(packet) {
			return true
		}
	}
	return false
}

func (pq *PriorityQueue) Remove(packet *packet.Packet) {
	for i, p := range pq.items {
		if p == packet {
			heap.Remove(pq, i)
			return
		}
	}
}

func (pq *PriorityQueue) Update(packet *packet.Packet, newPriority packet.Priority) {
	for i, p := range pq.items {
		if p == packet {
			p.Priority = newPriority
			heap.Fix(pq, i)
			return
		}
	}
}

func (pq *PriorityQueue) Size() int {
	return pq.Len()
}

/* FIXME: This should go to the queue package also
var err error
var ok bool
var hold [](*packet.Packet)
var p *packet.Packet
for {
	// Pull packets from the queue until we find one that can be sent
	if !w.queue.IsEmpty() {
		p = w.queue.PopPacket()
	}

	ok, err = w.sendCheck(p)
	if err != nil {
		w.log.Debug("sendCheck: %v", err)
		continue
	}
	if ok {
		w.send(p)
		// Return the packet to the pool for reuse
		w.pool.Put(p)
		break
	}
}
// Return the held packets to the queue
for _, p = range hold {
	w.queue.PushPacket(p)
}
hold = hold[:0]
*/

/*
	 FIXME: This should go to the queue package
		// FIXME: Move this to queue? Need to work out the best way to handle repeats. Probably in the queue
		// last    map[uint16]time.Time // Last time a packet was sent for each loco address (5ms minimum)

	func (w *Wavegen) SendPacketIdempotent(p *packet.Packet) {
		if w.queue.Contains(p) {
			// If the packet is already in the queue, do nothing
			return
		}
		w.SendPacket(p)
	}

	func (w *Wavegen) SendPacket(p *packet.Packet) {
		if p == nil || len(p.Bytes()) == 0 {
			w.log.Debug("invalid packet: %s", p)
			return
		}

		// Need to make sure the packet has a valid address to fulfill the minimum time between packets
		if p.Address == 0 {
			w.log.Debug("invalid packet address: %s", p)
			return
		}
		w.queue.PushPacket(p)

		/* FIXME: Need to figure out what clearResets does exactly/what purpose it serves
		clearResets();
		* /
	}
	func (w *Wavegen) SendMessage(data []byte, address uint16, repeats int, priority packet.Priority) {
		if len(data) == 0 {
			return
		}

		p := w.pool.Get().(*packet.Packet)
		p.Fill(data, address, priority, repeats)
		w.queue.PushPacket(p)

		/* FIXME: Need to figure out what clearResets does exactly/what purpose it serves
		clearResets();
		* /
	}
*/

/* FIXME: Cleanup
// Pass packets from the queue to the PIO state machine in order of priority
func (w *Wavegen) Run() {
	// The input format is an 8 bit number containing the number of bytes in the message,
	// followed by the data bytes. For example, the standard idle packet is 0x3FF00FF
	// 3 for the length, followed by 11111111 00000000 11111111
	// The message start bit, byte terminating bits, and the packet end bit are added automatically.
	// If the FIFO is empty the statemachine will send idle packets until stopped.

	var err error
	var ok bool
	var hold [](*packet.Packet)
	var p *packet.Packet
	for {
		// Send 'em (packets) if you got 'em
		if !w.queue.IsEmpty() {
			for {
				// Pull packets from the queue until we find one that can be sent
				if !w.queue.IsEmpty() {
					p = w.queue.PopPacket()
				} else {
					// If we run out of packets in the queue, send an idle packet
					p = w.NewPacket()
					p.Fill([]byte{0xFF, 0x00}, packet.Broadcast, packet.BestEffortPriority, 0)
				}

				ok, err = w.sendCheck(p)
				if err != nil {
					w.log.Debug("sendCheck: %v", err)
					continue
				}
				if ok {
					w.send(p)
					// Return the packet to the pool for reuse
					w.pool.Put(p)
					break
				}
			}
			// Return the held packets to the queue
			for _, p = range hold {
				w.queue.PushPacket(p)
			}
			hold = hold[:0]
		}
		time.Sleep(1 * time.Millisecond)
	}
}

// FIXME: This should go to the queue package

// FIXME: Remove this in favor of checking for < 2 locos on track?
func (w *Wavegen) sendCheck(p *packet.Packet) (bool, error) {
	if p.Address == packet.Broadcast {
		// Broadcast packets can always be sent
		return true, nil
	}

	// Check if the loco address is already in the last map
	if lastTime, ok := w.last[p.Address]; ok {
		// If the time since the last packet is less than 5ms, skip sending this packet
		if time.Since(lastTime) < 5*time.Millisecond {
			return false, nil
		}
	}

	// Update the last time for this loco address
	w.last[p.Address] = time.Now()
	return true, nil
}
*/
