//go:build rp

package main

import (
	"machine"
	"time"

	"github.com/mikesmitty/beacon-dcc/pkg/adc"
	"github.com/mikesmitty/beacon-dcc/pkg/dcc"
	"github.com/mikesmitty/beacon-dcc/pkg/dccex"
	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/motor"
	"github.com/mikesmitty/beacon-dcc/pkg/queue"
	"github.com/mikesmitty/beacon-dcc/pkg/serial"
	"github.com/mikesmitty/beacon-dcc/pkg/shared"
	"github.com/mikesmitty/beacon-dcc/pkg/topic"
	"github.com/mikesmitty/beacon-dcc/pkg/track"
	"github.com/mikesmitty/beacon-dcc/pkg/wavegen"
)

var (
	board      = "metro-rp2350"
	gitSHA     = ""
	shieldName = "EX-MotorShield8874"
	version    = ""
)

func main() {
	/* FIXME: Cleanup
	time.Sleep(3 * time.Second)

	led := machine.GPIO23
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.High()
	*/

	bus := event.NewEventBus()
	cl := bus.NewEventClient("main", "main", 1)

	dex := dccex.NewDCCEX(bus.NewEventClient("dccex", topic.BroadcastDex, 1)) // DCC-EX Command Parser
	dex.Subscribe(topic.ReceiveCmdSerial)

	defaultSerial := serial.NewSerial(machine.Serial, bus.NewEventClient("serial", topic.ReceiveCmdSerial))
	defaultSerial.Subscribe(topic.BroadcastDex, topic.BroadcastDebug)
	// defaultSerial.Update()

	// FIXME: Handle initialization of USBCDC if necessary
	// usbSerial := serial.NewSerial(bus, "usb", machine.USBCDC, topic.BroadcastDex, topic.BroadcastDebug)
	// usbSerial.Update()
	// defaultUART := c.RegisterSerial("uart", machine.DefaultUART, true)
	// defaultUART.Update()
	cl.Debug("Default serial initialized")

	q := queue.NewPriorityQueue(32, bus.NewEventClient("priorityqueue", topic.WavegenSend, 3))
	q.Subscribe(topic.WavegenQueue)

	w, err := InitWavegen(machine.GPIO22, machine.GPIO9, bus)
	if err != nil {
		panic(err)
	}
	cl.Debug("Wavegen initialized")

	profiles := motor.ShieldEX8874
	// profiles := motor.ArduinoMotorShieldRev3
	motors := InitMotor(profiles, bus)
	cl.Debug("Motor initialized")

	tracks := make(map[string]*track.Track)
	for id := range motors {
		tracks[id] = track.NewTrack(id, track.TrackModeMain, bus.NewEventClient("track"+id, topic.MotorControl+id, 1))
		tracks[id].Subscribe(topic.MotorControl + id)
	}

	board := shared.BoardInfo{
		Board:      board,
		GitSHA:     gitSHA,
		ShieldName: shieldName,
		Version:    version,
	}
	d := dcc.NewDCC(board, w, bus.NewEventClient("dcc", topic.BroadcastDex, 1))
	cl.Debug("DCC initialized")

	//
	// FIXME: Start converting again from here
	//

	// FIXME: Add all the non-tight loops

	go d.Run()

	for {
		for _, m := range motors {
			m.Update()
		}
		// runtime.Gosched()
		time.Sleep(100 * time.Microsecond) // FIXME: Cleanup - this should be a tight loop eventually
	}
}

func InitWavegen(signalPin, brakePin machine.Pin, bus *event.EventBus) (*wavegen.Wavegen, error) {
	signalPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	brakePin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	config := wavegen.WavegenConfig{
		Mode:      wavegen.NormalMode,
		SignalPin: signalPin,
		BrakePin:  brakePin,
	}

	w, err := wavegen.NewWavegen(config, bus.NewEventClient("wavegen", topic.WavegenQueue, 3))
	if err != nil {
		return nil, err
	}
	w.Subscribe(topic.WavegenSend)

	// Enable the wavegen
	w.Enable(true)

	return w, nil
}

func InitMotor(profiles map[string]motor.MotorShieldProfile, bus *event.EventBus) map[string]*motor.Motor {
	adcs := map[string]motor.ADC{
		"A": adc.NewADC(machine.ADC1),
		"B": adc.NewADC(machine.ADC2),
	}

	motors := make(map[string]*motor.Motor)
	for id, profile := range profiles {
		profile.ADC = adcs[id]
		motors[id] = motor.NewMotor(id, profile, bus.NewEventClient(id, topic.MotorStatus+id, 1))
		motors[id].Subscribe(topic.MotorControl + id)
	}

	// EX-MotorDriver8874
	// mA := motor.NewMotor("A", adc1, machine.GPIO3, machine.GPIO9, machine.GPIO45, false, 1.27, 5000, c)
	// mB := motor.NewMotor("B", adc2, machine.GPIO11, machine.GPIO8, machine.GPIO46, false, 1.27, 5000, c)
	// Arduino Motor Shield R3
	// mA := motor.NewMotor("A", adc1, machine.GPIO3, machine.GPIO9, machine.NoPin, false, 0.488, 1500, c)
	// mB := motor.NewMotor("B", adc2, machine.GPIO11, machine.GPIO8, machine.NoPin, false, 0.488, 1500, c)

	return motors
}
