package packet

import (
	"testing"
)

func TestNewPacket(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	packet := NewPacket(len(data) + 2)
	packet.Fill(data, Broadcast, 1, 1)
	packet.Encode() // This will add the checksum to the packet

	l := len(packet.data)
	if l != 4 {
		t.Errorf("Expected packet data length to be 4, got %d", l)
	}

	expectedChecksum := byte(0x00 ^ 0x01 ^ 0x02 ^ 0x03)
	if packet.data[l-1] != expectedChecksum {
		t.Errorf("Expected checksum to be %x, got %x", expectedChecksum, packet.data[l-1])
	}
}

// FIXME: Checksum is now only added when Encode is called
func TestPacketAddChecksum(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	packet := NewPacket(len(data) + 2)
	packet.Fill(data, Broadcast, 1, 1)
	expectedChecksum := byte(0x00 ^ 0x01 ^ 0x02 ^ 0x03)
	packet.Encode() // This will add the checksum to the packet

	l := len(packet.data)
	if l != 4 {
		t.Errorf("Expected packet data length to be 4, got %d", l)
	}

	if packet.data[l-1] != expectedChecksum {
		t.Errorf("Expected checksum to be %x, got %x", expectedChecksum, packet.data[l-1])
	}
}
