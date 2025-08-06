package dcc

import (
	"runtime"
	"sync"
	"time"

	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/packet"
	"github.com/mikesmitty/beacon-dcc/pkg/shared"
	"github.com/mikesmitty/beacon-dcc/pkg/wavegen"
)

/* FIXME: Cleanup
// This module is responsible for converting API calls into
// messages to be sent to the waveform generator.
// It has no visibility of the hardware, timers, interrupts
// nor of the waveform issues such as preambles, start bits checksums or cutouts.
//
// Nor should it have to deal with JMRI responsess other than the OK/FAIL
// or cv value returned. I will move that back to the JMRI interface later
//
// The interface to the waveform generator is narrowed down to merely:
//   Scheduling a message on the prog or main track using a function
//   Obtaining ACKs from the prog track using a function
//   There are no volatiles here.
*/

const (
	MaxLocos      = 1024
	MaxPacketSize = 32
)

const (
	cmdSetSpeed = 0x3F // 0b00111111

	shortAddressMax = 127
)

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

	loopState  LoopState
	state      map[uint16]LocoState
	stateMutex sync.RWMutex // TODO: Evaluate sync.Map for high concurrency
	wavegen    *wavegen.Wavegen

	*event.EventClient
}

func NewDCC(boardInfo shared.BoardInfo, wavegen *wavegen.Wavegen, cl *event.EventClient) *DCC {
	d := &DCC{
		BoardInfo:   boardInfo,
		EventClient: cl,

		wavegen: wavegen, // TODO: Make this an interface
	}

	return d
}

func (d *DCC) Run() {
	d.Diag("DCC-EX V-%s / %s / %s G-%s", d.Version, d.Board, d.ShieldName, d.GitSHA) // FIXME: Change hardcoded name?

	for {
		// Issue reminders for all locos, one type at a time, starting with throttle reminders, then functions by group
		// The messages are low priority and queued idempotently so there's no lower bound on the loop time, but we
		// don't want to spend more time than necessary checking the queue for duplicate packets.
		d.issueReminders()
		time.Sleep(5 * time.Millisecond)
	}
}

func (d *DCC) NewPacket(loco uint16) *packet.Packet {
	p := d.wavegen.NewPacket()
	if p == nil {
		d.Debug("Failed to create new packet for loco %d", loco)
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
	for loco := range d.state {
		d.sendReminderPackets(loco)
		runtime.Gosched() // Yield to allow other goroutines to run
	}
	d.loopState++
	if d.loopState >= LoopStateRestart {
		d.loopState = LoopStateSpeed
	}
}

// FIXME: Move this to queue package?
func (d *DCC) sendReminderPackets(loco uint16) {
	state, err := d.LocoState(loco)
	if err != nil {
		d.Debug("error getting loco %d state: %v", loco, err)
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

	p.Priority = packet.LowPriority
	p.Repeats = 0
	d.wavegen.SendPacketIdempotent(p)
}

func speedStep28(speed128 uint8) uint8 {
	if speed128 == 0 || speed128 == 1 {
		// Stop or emergency stop
		return speed128
	}
	// Convert 2-127 to 1-28
	speed28 := (speed128*10 + 36) / 46

	code28 := (speed28 + 3) / 2
	if (speed28 & 1) == 0 {
		code28 |= 0b00010000
	}
	//        Construct command byte from:
	//        command      speed    direction
	return 0b01000000 | code28 | (speed128&0x80)>>2
}
