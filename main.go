//go:build rp

package main

import (
	"machine"
	"time"

	"github.com/mikesmitty/beacon-dcc/pkg/wavegen"
)

func main() {
	time.Sleep(3 * time.Second)

	signalPin := machine.GPIO22
	signalPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	brakePin := machine.GPIO9
	brakePin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// w, err := wavegen.NewWavegen(0, signalPin, brakePin)
	_, err := wavegen.NewWavegen(0, signalPin, brakePin)
	if err != nil {
		panic(err)
	}
	println("Wavegen initialized")

	// state := brakePin.Get()
	// now := time.Now()
	// then := now
	// w.Enable(true) // FIXME: Cleanup
	for {
		// if state != brakePin.Get() {
		// 	state = brakePin.Get()
		// 	now = time.Now()
		// 	if state {
		// 		fmt.Printf("1: %d us\r\n", now.Sub(then).Microseconds())
		// 	} else {
		// 		fmt.Printf("0: %d us\r\n", now.Sub(then).Microseconds())
		// 	}
		// 	then = now
		// }
		// fmt.Printf("X: %d\r\n", w.GetX())
		// signalPin.High()
		// brakePin.High()
		// time.Sleep(100 * time.Microsecond)
		// brakePin.Low()
		// time.Sleep(100 * time.Microsecond)
	}
}
