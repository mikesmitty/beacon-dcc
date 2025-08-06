package event

import (
	"testing"
	"time"
)

// Mock EventBus for testing
type mockEventBus struct {
	subscribed   map[string]map[string]chan Event
	unsubscribed []string
	published    []struct {
		topic    string
		clientID string
		data     any
	}
}

func newMockEventBus() *mockEventBus {
	return &mockEventBus{
		subscribed: make(map[string]map[string]chan Event),
	}
}

func (m *mockEventBus) Subscribe(topic, clientID string, ch chan Event) {
	if m.subscribed[topic] == nil {
		m.subscribed[topic] = make(map[string]chan Event)
	}
	m.subscribed[topic][clientID] = ch
}

func (m *mockEventBus) Unsubscribe(topic, clientID string) {
	m.unsubscribed = append(m.unsubscribed, topic+":"+clientID)
	if m.subscribed[topic] != nil {
		delete(m.subscribed[topic], clientID)
	}
}

func (m *mockEventBus) Publish(topic, clientID string, data any) {
	m.published = append(m.published, struct {
		topic    string
		clientID string
		data     any
	}{topic, clientID, data})
}

func TestNewEventClient_DefaultBufferSize(t *testing.T) {
	eb := &EventBus{}
	client := eb.NewEventClient("id1", "topic1")
	if client.ClientID != "id1" {
		t.Errorf("expected ClientID 'id1', got %s", client.ClientID)
	}
	if client.DefaultPub != "topic1" {
		t.Errorf("expected DefaultPub 'topic1', got %s", client.DefaultPub)
	}
	if cap(client.Events) != defaultBufferSize {
		t.Errorf("expected buffer size %d, got %d", defaultBufferSize, cap(client.Events))
	}
}

func TestNewEventClient_CustomBufferSize(t *testing.T) {
	eb := &EventBus{}
	client := eb.NewEventClient("id2", "topic2", 5)
	if cap(client.Events) != 5 {
		t.Errorf("expected buffer size 5, got %d", cap(client.Events))
	}
}

func TestEventClient_SubscribeAndUnsubscribe(t *testing.T) {
	bus := newMockEventBus()
	client := &EventClient{
		ClientID:   "cid",
		Bus:        bus,
		Events:     make(chan Event, 2),
		DefaultPub: "def",
	}
	client.Subscribe("t1", "t2")
	if len(client.Topics) != 2 {
		t.Errorf("expected 2 topics, got %d", len(client.Topics))
	}
	if bus.subscribed["t1"]["cid"] == nil || bus.subscribed["t2"]["cid"] == nil {
		t.Error("client not subscribed to topics")
	}

	client.Unsubscribe("t1")
	if len(client.Topics) != 1 || client.Topics[0] != "t2" {
		t.Errorf("expected topics to be [t2], got %v", client.Topics)
	}
	if bus.subscribed["t1"]["cid"] != nil {
		t.Error("client should be unsubscribed from t1")
	}
}

func TestEventClient_UnsubscribeFromAll(t *testing.T) {
	bus := newMockEventBus()
	client := &EventClient{
		ClientID:   "cid",
		Bus:        bus,
		Events:     make(chan Event, 2),
		DefaultPub: "def",
		Topics:     []string{"t1", "t2"},
	}
	client.UnsubscribeFromAll()
	if len(client.Topics) != 0 {
		t.Errorf("expected no topics, got %v", client.Topics)
	}
	if len(bus.unsubscribed) != 2 {
		t.Errorf("expected 2 unsubscribed calls, got %d", len(bus.unsubscribed))
	}
}

func TestEventClient_Publish(t *testing.T) {
	bus := newMockEventBus()
	client := &EventClient{
		ClientID:   "cid",
		Bus:        bus,
		Events:     make(chan Event, 2),
		DefaultPub: "def",
	}
	client.Publish("data1")
	if len(bus.published) != 1 {
		t.Errorf("expected 1 publish, got %d", len(bus.published))
	}
	if bus.published[0].topic != "def" || bus.published[0].clientID != "cid" || bus.published[0].data != "data1" {
		t.Error("publish data mismatch")
	}
}

func TestEventClient_PublishTo(t *testing.T) {
	bus := newMockEventBus()
	client := &EventClient{
		ClientID:   "cid",
		Bus:        bus,
		Events:     make(chan Event, 2),
		DefaultPub: "def",
	}
	client.PublishTo("topicX", "dataX")
	if len(bus.published) != 1 {
		t.Errorf("expected 1 publish, got %d", len(bus.published))
	}
	if bus.published[0].topic != "topicX" || bus.published[0].clientID != "cid" || bus.published[0].data != "dataX" {
		t.Error("publishTo data mismatch")
	}
}

// Optional: Test concurrent publish/subscribe
func TestEventClient_ConcurrentPublishSubscribe(t *testing.T) {
	bus := newMockEventBus()
	client := &EventClient{
		ClientID:   "cid",
		Bus:        bus,
		Events:     make(chan Event, 10),
		DefaultPub: "def",
	}
	done := make(chan struct{})
	go func() {
		client.Subscribe("topic1")
		client.PublishTo("topic1", "payload")
		done <- struct{}{}
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("timeout in concurrent publish/subscribe")
	}
}
