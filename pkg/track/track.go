package track

import (
	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/topic"
)

type Track struct {
	id   string
	mode TrackMode

	*event.EventClient
}

func NewTrack(id string, mode TrackMode, cl *event.EventClient) *Track {
	t := &Track{
		id:          id,
		EventClient: cl,
		mode:        mode,
	}

	t.Subscribe(topic.TrackPowerOn)
	t.Subscribe(topic.TrackPowerOff)
	t.Debug("Track %s initialized with mode %s", id, mode)

	return t
}

func (t *Track) Update() {
	select {
	case evt := <-t.Events:
		switch msg := evt.Data.(type) {
		case string:
			t.handleStringEvent(evt, msg)
		default:
			t.Debug("Received unknown event type: %T", evt.Data)
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
		t.PublishTo(topic.MotorControl+t.id, "on")
	case topic.TrackPowerOff:
		t.PublishTo(topic.MotorControl+t.id, "off")
	}
}

func (t *Track) SetMode(mode TrackMode) {
	t.mode = mode
	t.Debug("Track mode set to %s", mode)
}
