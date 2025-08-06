package serial

import (
	"bytes"
	"testing"

	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/topic"
)

// MockSerialer implements Serialer for testing
type MockSerialer struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
}

func (m *MockSerialer) WriteByte(c byte) error {
	return m.writeBuf.WriteByte(c)
}
func (m *MockSerialer) Write(data []byte) (n int, err error) {
	return m.writeBuf.Write(data)
}
func (m *MockSerialer) Buffered() int {
	return m.readBuf.Len()
}
func (m *MockSerialer) ReadByte() (byte, error) {
	return m.readBuf.ReadByte()
}

func TestReadCommand_ValidCommand(t *testing.T) {
	eb := event.NewEventBus()
	mockSerial := &MockSerialer{
		readBuf:  bytes.NewBufferString("<abc>"),
		writeBuf: new(bytes.Buffer),
	}
	s := NewSerial(mockSerial, eb.NewEventClient("test", topic.ReceiveCmdSerial))

	cl := eb.NewEventClient("test", "")
	cl.Subscribe(topic.ReceiveCmdSerial)

	s.ReadCommand()

	select {
	case evt := <-cl.Events:
		buf, ok := evt.Data.(*bytes.Buffer)
		if !ok {
			t.Fatal("event data is not *bytes.Buffer")
		}
		if buf.String() != "abc" {
			t.Errorf("expected 'abc', got '%s'", buf.String())
		}
	default:
		t.Fatal("no event received")
	}
}

func TestReadCommand_JunkBeforeStart(t *testing.T) {
	eb := event.NewEventBus()
	mockSerial := &MockSerialer{
		readBuf:  bytes.NewBufferString("junk<cmd>"),
		writeBuf: new(bytes.Buffer),
	}
	s := NewSerial(mockSerial, eb.NewEventClient("test", topic.ReceiveCmdSerial))

	cl := eb.NewEventClient("test", "")
	cl.Subscribe(topic.ReceiveCmdSerial)

	s.ReadCommand()

	select {
	case evt := <-cl.Events:
		buf, ok := evt.Data.(*bytes.Buffer)
		if !ok {
			t.Fatal("event data is not *bytes.Buffer")
		}
		if buf.String() != "cmd" {
			t.Errorf("expected 'cmd', got '%s'", buf.String())
		}
	default:
		t.Fatal("no event received")
	}
}

func TestReadCommand_IncompleteCommand(t *testing.T) {
	eb := event.NewEventBus()
	mockSerial := &MockSerialer{
		readBuf:  bytes.NewBufferString("<incomplete"),
		writeBuf: new(bytes.Buffer),
	}
	s := NewSerial(mockSerial, eb.NewEventClient("test", topic.ReceiveCmdSerial))

	cl := eb.NewEventClient("test", "")
	cl.Subscribe(topic.ReceiveCmdSerial)

	s.ReadCommand()

	select {
	case <-cl.Events:
		t.Fatal("should not receive event for incomplete command")
	default:
		// success
	}
}

func TestRun_ClientsDex(t *testing.T) {
	eb := event.NewEventBus()
	mockSerial := &MockSerialer{
		readBuf:  new(bytes.Buffer),
		writeBuf: new(bytes.Buffer),
	}
	s := NewSerial(mockSerial, eb.NewEventClient("test", topic.ReceiveCmdSerial))
	s.Subscribe(topic.BroadcastDex)

	msg := bytes.NewBufferString("hello")
	eb.Publish(topic.BroadcastDex, "test", msg)

	s.Update()

	got := mockSerial.writeBuf.String()
	if got == "" || got[:5] != "hello" {
		t.Errorf("expected 'hello', got '%s'", got)
	}
}
