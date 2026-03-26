package queue

import (
	"container/heap"
	"testing"

	"github.com/mikesmitty/beacon-dcc/pkg/packet"
)

func TestPriorityQueue(t *testing.T) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	packet1 := packet.NewPacket(32)
	packet2 := packet.NewPacket(32)
	packet3 := packet.NewPacket(32)
	packet1.Fill([]byte{0x01}, packet.Broadcast, 10, 1)
	packet2.Fill([]byte{0x02}, packet.Broadcast, 20, 1)
	packet3.Fill([]byte{0x03}, packet.Broadcast, 10, 0)

	pq.PushPacket(packet1)
	pq.PushPacket(packet2)
	pq.PushPacket(packet3)

	if pq.Len() != 3 {
		t.Errorf("Expected length 3, got %d", pq.Len())
	}

	peek := pq.Peek().Bytes()
	if peek[0] != 0x02 {
		t.Errorf("Expected highest priority packet to be 0x02, got %x", peek[0])
	}

	popped := pq.PopPacket()
	if popped.Bytes()[0] != 0x02 || popped.Repeats != 0 {
		t.Errorf("Expected to pop packet with data 0x02 and repeats 0, got %x with repeats %d", popped.Bytes()[0], popped.Repeats)
	}

	if pq.Len() != 3 {
		t.Errorf("Expected length 3 after pop, got %d", pq.Len())
	}

	pq.PopPacket()
	if pq.Len() != 2 {
		t.Errorf("Expected length 2 after pop, got %d", pq.Len())
	}
}

func TestPopOrder(t *testing.T) {
	pq := &PriorityQueue{}
	heap.Init(pq)

	packet1 := packet.NewPacket(32)
	packet2 := packet.NewPacket(32)
	packet1.Fill([]byte{0x01}, packet.Broadcast, 10, 1)
	packet2.Fill([]byte{0x02}, packet.Broadcast, 10, 1)

	pq.PushPacket(packet1)
	pq.PushPacket(packet2)

	popped1 := pq.PopPacket().Bytes()
	popped2 := pq.PopPacket().Bytes()
	if popped1[0] == popped2[0] {
		t.Errorf("Expected repeated packets to be round-robin, instead got matching packets: % x", popped1)
	}
}
