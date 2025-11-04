package track

import (
	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/topic"
)

type Track struct {
	id   string
	mode TrackMode

	Event *event.EventClient
}

func NewTrack(id string, mode TrackMode, cl *event.EventClient) *Track {
	t := &Track{
		id:    id,
		Event: cl,
		mode:  mode,
	}

	t.Event.Subscribe(topic.TrackPowerOn)
	t.Event.Subscribe(topic.TrackPowerOff)
	t.Event.Debug("Track %s initialized with mode %s", id, mode)

	return t
}

func (t *Track) Update() {
	select {
	case evt := <-t.Event.Receive:
		switch msg := evt.Data.(type) {
		case string:
			t.handleStringEvent(evt, msg)
		default:
			t.Event.Debug("Received unknown event type: %T", evt.Data)
		}
	default:
		// No event to process
	}
}

func (t *Track) handleStringEvent(evt event.Event, msg string) {
	if msg != t.id && msg != "all" {
		return // Ignore events not for this track
	}
	switch evt.Topic {
	case topic.TrackPowerOn:
		t.Event.PublishTo(topic.MotorControl+t.id, "on")
	case topic.TrackPowerOff:
		t.Event.PublishTo(topic.MotorControl+t.id, "off")
	}
}

func (t *Track) SetMode(mode TrackMode) {
	t.mode = mode
	t.Event.Debug("Track mode set to %s", mode)
}
