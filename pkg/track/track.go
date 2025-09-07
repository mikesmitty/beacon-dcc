package track

import (
	"errors"

	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/topic"
)

type Track struct {
	driver Driver
	id     string
	mode   TrackMode

	Event *event.EventClient
}

func NewTrack(id string, mode TrackMode, driver Driver, cl *event.EventClient) (*Track, error) {
	if driver == nil {
		return nil, errors.New("Track driver cannot be nil")
	}
	t := &Track{
		id:     id,
		Event:  cl,
		mode:   mode,
		driver: driver,
	}

	driver.SetEventClient(cl)
	if err := driver.Init(id); err != nil {
		return nil, err
	}

	t.Event.Subscribe(topic.TrackModeJoin)
	t.Event.Subscribe(topic.TrackModeUnjoin)
	t.Event.Subscribe(topic.TrackPowerOn)
	t.Event.Subscribe(topic.TrackPowerOff)
	t.Event.Debug("Track %s initialized with mode %s", id, mode)

	return t, nil
}

func (t *Track) Update() {
	t.update()
	t.driver.Update()
}

func (t *Track) update() {
	select {
	case evt := <-t.Event.Receive:
		switch msg := evt.Data.(type) {
		case string:
			if !t.isValidString(msg) {
				return
			}
		case TrackMode:
			if !t.isValidTrackMode(evt, msg) {
				return
			}
		default:
			t.Event.Debug("Received unknown event type: %T", evt.Data)
			return
		}
		t.handleEvent(evt)
	default:
		// No event to process
	}
}

func (t *Track) isValidTrackMode(evt event.Event, mode TrackMode) bool {
	switch evt.Topic {
	case topic.TrackModeJoin, topic.TrackModeUnjoin:
		// FIXME: Need to check for track ID?
		return t.mode.IsProg()
	default:
		return t.mode.Matches(mode)
	}
}

func (t *Track) isValidString(msg string) bool {
	switch msg {
	case "all", t.id, t.mode.String():
		// Continue handling
		return true
	default:
		// Ignore events not for this track
		return false
	}
}

func (t *Track) handleEvent(evt event.Event) {
	switch evt.Topic {
	case topic.TrackModeJoin:
		t.AddMode(TrackModeMain)
	case topic.TrackModeUnjoin:
		t.RemoveMode(TrackModeMain)
	case topic.TrackPowerOn:
		t.driver.SetPower(true)
	case topic.TrackPowerOff:
		t.driver.SetPower(false)
	}
}

func (t *Track) SetMode(mode TrackMode) {
	t.mode = mode
	t.Event.Debug("Track mode set to %s", mode)
}

func (t *Track) AddMode(mode TrackMode) {
	t.mode |= mode
	t.Event.Debug("Track mode added: %s", mode)
}

func (t *Track) RemoveMode(mode TrackMode) {
	t.mode &^= mode
	t.Event.Debug("Track mode removed: %s", mode)
}
