package dcc

import (
	"fmt"

	"github.com/mikesmitty/beacon-dcc/pkg/packet"
)

type AccOnOff byte

const (
	AccOff  AccOnOff = iota // 0
	AccOn                   // 1
	AccBoth                 // 2
)

func (d *DCC) setAccessory(address uint16, port byte, direction bool, onOff AccOnOff) {
	// NMRA S-9.2.1 - 2.4.1 Basic Accessory Decoder Packet Format
	// By convention, R = 0 means diverging, direction of travel to the left, or signal to stop
	// and R = 1 means normal, direction of travel to the right, or signal to proceed
	d.Debug("setAccessory(%d, %d, %t, %d)", address, port, direction, onOff)

	// The Basic Accessory Decoder format only supports 9-bit addresses and 4 ports.
	if address > 511 || port > 3 {
		return
	}

	pOn := d.wavegen.NewPacket()
	pOff := d.wavegen.NewPacket()
	if pOn == nil || pOff == nil {
		d.Debug("Failed to create new packet for accessory command")
		return
	}

	// First byte format: 10AAAAAA, where AAAAAA represents the least significant address bits.
	// Second byte format: 1AAACPPR, where C is on/off, PP is the port (0-3) and R is the direction.
	lsb := byte(address % 64)
	msb := ^(byte(address / 64)) & 0b111 // The most significant address bits are inverted
	onBit := byte(1)
	dirBit := byte(0)
	if direction {
		dirBit = 0x01 // Set direction bit to 1 if direction is true
	}

	pOn.AddByte(0x80 | lsb)
	pOn.AddByte(0x80 | (msb << 4) | (onBit << 3) | (port << 1) | dirBit)

	pOff.AddByte(0x80 | lsb)
	pOff.AddByte(0x80 | (msb << 4) | (port << 1) | dirBit)

	// Set the on command priority one higher to ensure its repeats are all processed before the off command.
	// With equal priority they would be sent in a round-robin order.
	pOn.Priority = packet.NormalPriority + 1
	pOn.Repeats = 3

	pOff.Priority = packet.NormalPriority
	pOff.Repeats = 3

	// AccOn || AccBoth
	if onOff != AccOff {
		d.wavegen.SendPacket(pOn)
		// FIXME: Cleanup
		// #if defined(EXRAIL_ACTIVE)
		//     RMFT2::activateEvent(address<<2|port, direction);
		// #endif
	} else {
		// Discard the unused packet to return it to the pool
		d.wavegen.DiscardPacket(pOn)
	}

	// AccOff || AccBoth
	if onOff != AccOn {
		d.wavegen.SendPacket(pOff)
	} else {
		// Discard the unused packet to return it to the pool
		d.wavegen.DiscardPacket(pOff)
	}
}

func (d *DCC) setExtendedAccessory(address uint16, value byte, repeats int) error {
	if address > 2044 || address < 1 || value > 31 {
		return fmt.Errorf("invalid address or value: address=%d, value=%d", address, value)
	}

	p := d.wavegen.NewPacket()
	if p == nil {
		return fmt.Errorf("failed to create new packet for extended accessory command")
	}

	// NMRA S-9.2.1 - 2.4.2 Extended Accessory Decoder Control Packet Format
	//   By convention, the first address, known to the user as address “1” starts at:
	//   {preamble} 0 10000001 0 01110001 0 XXXXXXXX 0 EEEEEEEE 1
	//
	// This corresponds to 4 (000 00000100), so we adjust the address by +3 here
	// Zero and negative addresses don't exist, hence the usable range of 1-2044
	address += 3
	addrHigh := byte(address >> 8) // Bits A10-A8
	addrLow := byte(address & 0xFF)

	// Address bits A7-A2
	p.AddByte(0x80 | (addrLow >> 2))
	// A10-A8 - By convention these bits (bits 4 to 6 of the second byte) are in ones’ complement. (15)
	// (15) - The ones’ complement form is where every bit is inverted [...]
	// A10-A8 (inverted), A1-A0, static 1
	p.AddByte(^(addrHigh << 4) | ((addrLow & 0x3) << 1) | 0x1)
	p.AddByte(value)

	p.Priority = packet.NormalPriority
	p.Repeats = repeats
	d.wavegen.SendPacket(p)
	return nil
}
