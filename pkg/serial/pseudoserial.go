package serial

import (
	"io"
	"machine"
	"os"
	"time"
)

var _ Serialer = (*PseudoSerial)(nil)

type PseudoSerial struct {
	ch chan byte
}

func NewPseudoSerial() *PseudoSerial {
	p := &PseudoSerial{
		ch: make(chan byte, 256),
	}
	// Read from stdin in a dedicated goroutine to avoid blocking
	// the serial update loop and to eliminate data races.
	go p.reader()

	return p
}

func (p *PseudoSerial) reader() {
	buf := make([]byte, 64)
	for {
		n, err := os.Stdin.Read(buf)
		for i := 0; i < n; i++ {
			select {
			case p.ch <- buf[i]:
			default:
				// Channel full, drop byte
			}
		}
		if n == 0 || err != nil {
			// Yield to other goroutines when no data is available
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func (p *PseudoSerial) Buffered() int {
	return len(p.ch)
}

func (p *PseudoSerial) ReadByte() (byte, error) {
	select {
	case b := <-p.ch:
		return b, nil
	default:
		return 0, io.EOF
	}
}

func (p *PseudoSerial) WriteByte(c byte) error {
	_, err := os.Stdout.Write([]byte{c})
	return err
}

func (p *PseudoSerial) Write(data []byte) (n int, err error) {
	return os.Stdout.Write(data)
}

func (p *PseudoSerial) Configure(config machine.UARTConfig) error {
	// PseudoSerial does not require configuration
	return nil
}
