package track

import "github.com/mikesmitty/beacon-dcc/pkg/event"

type Driver interface {
	Init(trackId string) error
	SetEventClient(*event.EventClient)
	Update()

	Power() (bool, error)
	SetPower(bool) error
}
