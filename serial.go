//go:build rp

package main

import (
	"machine"
	"machine/usb/cdc"
	"reflect"

	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/serial"
)

type Serials struct {
	serials map[string]*serial.Serial
}

func InitSerials(bus *event.EventBus, pub string, subs ...string) *Serials {
	// We can't treat the default TinyGo serial device as a regular serial device so we pretend
	ptty := serial.NewPseudoSerial()
	serialOptions := map[string]Serialer{
		"serial": ptty,
	}
	// FIXME: Not working in -serial=uart mode
	// Always enable both USB and default UART
	if reflect.DeepEqual(machine.Serial, machine.USBCDC) {
		serialOptions["uart"] = machine.DefaultUART
	} else {
		usbcdc := machine.USBCDC
		if usbcdc == nil {
			usbcdc = cdc.New()
		}
		serialOptions["usb"] = usbcdc
	}

	const bufSize = 16
	serials := make(map[string]*serial.Serial)
	for id, s := range serialOptions {
		s.Configure(machine.UARTConfig{})
		serials[id] = serial.NewSerial(s, bus.NewEventClient(id, pub, bufSize))
		serials[id].Event.Subscribe(subs...)
	}
	return &Serials{serials: serials}
}

func (s *Serials) Update() {
	for _, s := range s.serials {
		s.Update()
	}
}

type Serialer interface {
	WriteByte(c byte) error
	Write(data []byte) (n int, err error)
	Buffered() int
	ReadByte() (byte, error)
	Configure(config machine.UARTConfig) error
}
