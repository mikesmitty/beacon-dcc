package event

import (
	"sync"
)

type Event struct {
	Topic    string
	ClientID string
	Data     any
}

type EventBus struct {
	subscribers map[string]map[string]chan Event
	mux         *sync.Mutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		mux:         &sync.Mutex{},
		subscribers: make(map[string]map[string]chan Event),
	}
}

func (eb *EventBus) Subscribe(topic string, clientId string, ch chan Event) {
	eb.mux.Lock()
	defer func() { // FIXME: Cleanup
		if r := recover(); r != nil {
			println("Recovered from panic in Subscribe:")
		}
	}()

	// Add the channel to the list of subscribers for this topic.
	if eb.subscribers[topic] == nil {
		eb.subscribers[topic] = make(map[string]chan Event)
	}
	eb.subscribers[topic][clientId] = ch
	eb.mux.Unlock()
}

func (eb *EventBus) Unsubscribe(topic string, clientId string) {
	eb.mux.Lock()
	defer func() { // FIXME: Cleanup
		if r := recover(); r != nil {
			println("Recovered from panic in Unsubscribe:")
		}
	}()

	if subs, ok := eb.subscribers[topic]; ok {
		delete(subs, clientId)
		// If the topic has no more subscribers, remove it too
		if len(subs) == 0 {
			delete(eb.subscribers, topic)
		}
	}
	eb.mux.Unlock()
}

func (eb *EventBus) Publish(topic, clientId string, data any) {
	eb.mux.Lock()
	defer func() { // FIXME: Cleanup
		if r := recover(); r != nil {
			println("Recovered from panic in Unsubscribe:")
		}
	}()

	if subscribers, ok := eb.subscribers[topic]; ok {
		event := Event{Topic: topic, ClientID: clientId, Data: data}

		for _, ch := range subscribers {
			select {
			case ch <- event:
			default:
				// The subscriber's channel was full. The event is dropped.
				// TODO: Consider logging this
			}
		}
	}
	eb.mux.Unlock()
}

func (eb *EventBus) SubscriberCount(topic string) int {
	eb.mux.Lock()
	defer func() { // FIXME: Cleanup
		if r := recover(); r != nil {
			println("Recovered from panic in SubscriberCount:")
		}
	}()
	subs := len(eb.subscribers[topic])
	eb.mux.Unlock()
	return subs
}
