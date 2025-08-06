package serial

import (
	"bytes"

	"github.com/mikesmitty/beacon-dcc/pkg/event"
)

type Serial struct {
	buf  *bytes.Buffer
	uart Serialer

	*event.EventClient
}

type Serialer interface {
	WriteByte(c byte) error
	Write(data []byte) (n int, err error)
	Buffered() int
	ReadByte() (byte, error)
}

func NewSerial(uart Serialer, cl *event.EventClient) *Serial {
	s := &Serial{
		buf:  new(bytes.Buffer),
		uart: uart,

		EventClient: cl,
	}

	if s.uart == nil {
		panic("UART not configured")
	}

	return s
}

func (s *Serial) Update() {
	select {
	case evt := <-s.Events:
		msg, ok := evt.Data.(*bytes.Buffer)
		if !ok {
			break
		}

		s.uart.Write(msg.Bytes())
		s.uart.Write([]byte("\r\n")) // FIXME: \r needed?

	default:
		s.ReadCommand()
	}
}

func (s *Serial) ReadCommand() {
	if s.uart == nil {
		return
	}

	for s.uart.Buffered() > 0 {
		b, err := s.uart.ReadByte()
		if err == nil {
			s.buf.WriteByte(b)
		}
	}

	for s.buf.Len() > 0 {
		data := s.buf.Bytes()

		start := bytes.IndexByte(data, '<')
		if start == -1 {
			// No start char '<' found, clear the cmdBuffer
			s.buf.Reset()
			break
		}

		if start > 0 {
			// Overwrite any junk data before the start of the command
			s.buf.Next(start)
			data = s.buf.Bytes()
		}

		end := bytes.IndexByte(data, '>')
		if end == -1 {
			// We have an incomplete command. Wait for more data.
			break
		}

		// Command is trimmed of the start '<' and end '>' before sending
		cmdBuf := new(bytes.Buffer)
		command := data[1:end]
		cmdBuf.Write(command)
		s.Publish(cmdBuf)

		// Reset to break the loop and be ready to start again
		s.buf.Reset()
	}
}
