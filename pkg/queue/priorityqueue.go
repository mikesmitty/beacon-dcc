package queue

import (
	"container/heap"
	"time"

	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/packet"
)

type PriorityQueue struct {
	items []*packet.Packet

	Event *event.EventClient
}

var _ heap.Interface = (*PriorityQueue)(nil)

func NewPriorityQueue(size int, eventClient *event.EventClient) PriorityQueue {
	q := PriorityQueue{
		items: make([]*packet.Packet, 0, size),
		Event: eventClient,
	}
	return q
}

func (pq *PriorityQueue) Loop() {
	for {
		select {
		case evt := <-pq.Event.Receive:
			switch p := evt.Data.(type) {
			case *packet.Packet:
				if p == nil {
					return
				}
				if !pq.Contains(p) {
					pq.PushPacket(p)
				}
			default:
				pq.Event.Debug("Received unknown event type: %T", evt.Data)
			}
		default:
			// No event to process
		}

		if !pq.IsEmpty() {
			pq.Event.Publish(pq.PopPacket())
		}
		time.Sleep(1 * time.Millisecond)
	}
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
	if packet == nil {
		pq.Event.Debug("Cannot push nil packet")
		return
	}
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

func (pq *PriorityQueue) UpdatePriority(packet *packet.Packet, newPriority packet.Priority) {
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
