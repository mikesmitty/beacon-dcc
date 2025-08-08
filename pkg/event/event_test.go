package event

import (
	"testing"
	"time"
)

func TestSubscribeAndPublish(t *testing.T) {
	bus := NewEventBus()
	topic := "testTopic"
	clientID := "client1"
	ch := make(chan Event, 1)

	bus.Subscribe(topic, clientID, ch)
	bus.Publish(topic, clientID, "hello")

	select {
	case evt := <-ch:
		if evt.Topic != topic || evt.ClientID != clientID || evt.Data != "hello" {
			t.Errorf("unexpected event: %+v", evt)
		}
	default:
		t.Error("expected event, got none")
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := NewEventBus()
	topic := "testTopic"
	clientID := "client1"
	ch := make(chan Event, 1)

	bus.Subscribe(topic, clientID, ch)
	bus.Unsubscribe(topic, clientID)
	bus.Publish(topic, clientID, "should not be received")

	select {
	case <-ch:
		t.Error("received event after unsubscribe")
	case <-time.After(50 * time.Millisecond):
		// success, no event received
	}
}

func TestPublishNoSubscribers(t *testing.T) {
	bus := NewEventBus()
	// Should not panic or block
	bus.Publish("noTopic", "noClient", "data")
}

func TestMultipleSubscribers(t *testing.T) {
	bus := NewEventBus()
	topic := "topic"
	ch1 := make(chan Event, 1)
	ch2 := make(chan Event, 1)

	bus.Subscribe(topic, "client1", ch1)
	bus.Subscribe(topic, "client2", ch2)
	bus.Publish(topic, "client1", "data")

	received := 0
	for _, ch := range []chan Event{ch1, ch2} {
		select {
		case evt := <-ch:
			if evt.Data != "data" {
				t.Errorf("unexpected data: %v", evt.Data)
			}
			received++
		default:
			t.Errorf("expected event on channel")
		}
	}
	if received != 2 {
		t.Errorf("expected 2 events, got %d", received)
	}
}

func TestChannelFullDropsEvent(t *testing.T) {
	bus := NewEventBus()
	topic := "topic"
	ch := make(chan Event, 1)
	bus.Subscribe(topic, "client1", ch)

	// Fill the channel
	ch <- Event{Topic: topic, ClientID: "client1", Data: "filled"}

	// This publish should be dropped
	bus.Publish(topic, "client1", "dropped")

	// Only the original event should be in the channel
	select {
	case evt := <-ch:
		if evt.Data != "filled" {
			t.Errorf("unexpected event data: %v", evt.Data)
		}
	default:
		t.Error("expected event in channel")
	}
}
