package dcc

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/packet"
	"github.com/mikesmitty/beacon-dcc/pkg/shared"
	"github.com/mikesmitty/beacon-dcc/pkg/topic"
)

const (
	MaxLocos      = 1024
	MaxPacketSize = 32
)

const (
	cmdSetSpeed = 0x3F // 0b00111111

	shortAddressMax = 127
)

type WaveGenerator interface {
	Enable(enabled bool)
	IdlePacket() *packet.Packet
}

type SpeedMode uint8

const (
	SpeedMode14  SpeedMode = 14
	SpeedMode28  SpeedMode = 28
	SpeedMode128 SpeedMode = 128
)

type LoopState uint8

const (
	LoopStateSpeed LoopState = iota
	LoopStateFnGroup1
	LoopStateFnGroup2
	LoopStateFnGroup3
	LoopStateFnGroup4
	LoopStateFnGroup5
	LoopStateRestart
)

type DCC struct {
	shared.BoardInfo

	loopState LoopState
	state     map[uint16]LocoState
	mu        *sync.RWMutex
	wavegen   WaveGenerator
	pool      *packet.PacketPool

	Event *event.EventClient
}

func NewDCC(boardInfo shared.BoardInfo, wavegen WaveGenerator, pool *packet.PacketPool, cl *event.EventClient) *DCC {
	d := &DCC{
		BoardInfo: boardInfo,
		Event:     cl,

		pool:    pool,
		state:   make(map[uint16]LocoState),
		mu:      &sync.RWMutex{},
		wavegen: wavegen,
	}

	d.Event.Diag("Beacon-DCC V-%s / %s / %s G-%s", d.Version, d.Board, d.ShieldName, d.GitSHA)

	return d
}

func (d *DCC) Update() {
	// Issue reminders for all locos, one type at a time, starting with throttle reminders, then functions by group
	// The messages are low priority and queued idempotently so there's no lower bound on the loop time, but we
	// don't want to spend more time than necessary checking the queue for duplicate packets.
	d.issueReminders()
}

func (d *DCC) NewPacket(loco uint16) *packet.Packet {
	p := d.pool.NewPacket()
	if p == nil {
		d.Event.Debug("Failed to create new packet for loco %d", loco)
		return nil
	}

	if loco > shortAddressMax {
		// Convert train number to long address format
		p.AddByte(0xC0 | byte(loco>>8))
	}
	p.AddByte(byte(loco))
	return p
}

func (d *DCC) MotorShieldName() string {
	return d.ShieldName
}

func (d *DCC) issueReminders() {
	d.mu.RLock()
	locos := make([]uint16, 0, len(d.state))
	for loco := range d.state {
		locos = append(locos, loco)
	}
	d.mu.RUnlock()

	for _, loco := range locos {
		d.sendReminderPackets(loco)
	}
	d.loopState++
	if d.loopState >= LoopStateRestart {
		d.loopState = LoopStateSpeed
	}
}

func (d *DCC) sendReminderPackets(loco uint16) {
	state, err := d.LocoState(loco)
	if err != nil {
		d.Event.Debug("error getting loco %d state: %v", loco, err)
		return
	}

	repeats := int(0)
	var fnByte byte
	var p *packet.Packet
	switch d.loopState {
	case LoopStateSpeed:
		p = d.setThrottlePacket(loco, state.SpeedStep)
	case LoopStateFnGroup1:
		if state.GroupFlags&fnGroup1 != 0 {
			fnByte = 0x80 | (byte(state.Functions>>1) & 0x0F) | (byte(state.Functions&0x01) << 4)
			p = d.setFunctionPacket(loco, repeats, fnByte) // 100D DDDD
		}
	case LoopStateFnGroup2:
		if state.GroupFlags&fnGroup2 != 0 {
			fnByte = 0xB0 | (byte(state.Functions>>5) & 0x0F)
			p = d.setFunctionPacket(loco, repeats, fnByte) // 1011 DDDD
		}
	case LoopStateFnGroup3:
		if state.GroupFlags&fnGroup3 != 0 {
			fnByte = 0xA0 | (byte(state.Functions>>9) & 0x0F)
			p = d.setFunctionPacket(loco, repeats, fnByte) // 1010 DDDD
		}
	case LoopStateFnGroup4:
		if state.GroupFlags&fnGroup4 != 0 {
			fnByte = byte(state.Functions >> 13)
			p = d.setFunctionPacket(loco, repeats, 0xDE, fnByte)
		}
	case LoopStateFnGroup5:
		if state.GroupFlags&fnGroup5 != 0 {
			fnByte = byte(state.Functions >> 21)
			p = d.setFunctionPacket(loco, repeats, 0xDF, fnByte)
		}
	case LoopStateRestart:
		// Reset loop state to speed reminder
		d.loopState = LoopStateSpeed
	}

	if p != nil {
		p.Priority = packet.LowPriority
		p.Repeats = 0
		d.Event.PublishTo(topic.WavegenQueue, p)
	}
}

func (d *DCC) Broadcastf(input string, args ...any) {
	buf := bytes.NewBuffer(fmt.Appendf(nil, input, args...))
	d.Event.Publish(buf)
}

func (d *DCC) Broadcast(input string) {
	buf := bytes.NewBufferString(input)
	d.Event.Publish(buf)
}

func speedStep28(speedStep uint8) uint8 {
	speed := speedStep & 0x7F
	direction := (speedStep & 0x80) != 0

	var code28 uint8
	if speed == 0 || speed == 1 {
		// Stop or emergency stop
		code28 = speed
	} else {
		// Convert 2-127 to 1-28
		speed28 := (speed*10 + 36) / 46

		code28 = (speed28 + 3) / 2
		if (speed28 & 1) == 0 {
			code28 |= 0b00010000
		}
	}
	// Construct command byte from: 01 Direction C Speed
	res := 0b01000000 | code28
	if direction {
		res |= 0b00100000
	}
	return res
}
