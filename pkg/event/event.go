package event

import (
	"maps"
	"slices"
	"sync"
)

type Event struct {
	Topic    string
	ClientID string
	Data     any
}

type EventBus struct {
	mu   sync.Mutex
	subs map[string]map[string]chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{
		subs: make(map[string]map[string]chan Event),
	}
}

func (eb *EventBus) Subscribe(topic string, clientId string, ch chan Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Add the channel to the list of subscribers for this topic.
	if eb.subs[topic] == nil {
		eb.subs[topic] = make(map[string]chan Event)
	}
	eb.subs[topic][clientId] = ch
}

func (eb *EventBus) Unsubscribe(topic string, clientId string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if subs, ok := eb.subs[topic]; ok {
		delete(subs, clientId)
		// If the topic has no more subscribers, remove it too
		if len(subs) == 0 {
			delete(eb.subs, topic)
		}
	}
}

func (eb *EventBus) Publish(topic, clientId string, data any) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if subscribers, ok := eb.subs[topic]; ok {
		event := Event{Topic: topic, ClientID: clientId, Data: data}

		for _, ch := range subscribers {
			select {
			case ch <- event:
			default:
				// The subscriber's channel was full. The event is dropped.
			}
		}
	}
}

func (eb *EventBus) Subscribers(topic string) []string {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	return slices.Sorted(maps.Keys(eb.subs[topic]))
}

func (eb *EventBus) SubscriberCount(topic string) int {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	return len(eb.subs[topic])
}
