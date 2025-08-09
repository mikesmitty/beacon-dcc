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
	"github.com/mikesmitty/beacon-dcc/pkg/packet"
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
	for {
		println("Starting loop...")
		loop() // FIXME: Cleanup
	}
}

func loop() {
	// defer func() { // FIXME: Cleanup
	// 	if r := recover(); r != nil {
	// 		println("Recovered from panic")
	// 	}
	// }()
	/* FIXME: Cleanup
	time.Sleep(3 * time.Second)

	led := machine.GPIO23
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})
	led.High()
	*/

	bus := event.NewEventBus()
	cl := bus.NewEventClient("main", "main")

	dex := dccex.NewDCCEX(bus.NewEventClient("dccex", topic.BroadcastDex))
	dex.Event.Subscribe(topic.ReceiveCmdSerial)

	serialOptions := map[string]Serialer{
		// "serial": machine.Serial,
		// FIXME: Handle initialization of USBCDC if necessary
		// "usb":    machine.USBCDC,
		"uart": machine.DefaultUART,
	}

	serials := make(map[string]*serial.Serial)
	for id, s := range serialOptions {
		s.Configure(machine.UARTConfig{})
		serials[id] = serial.NewSerial(s, bus.NewEventClient(id, topic.ReceiveCmdSerial))
		serials[id].Event.Subscribe(topic.BroadcastDex, topic.BroadcastDebug)
	}

	go func() {
		for {
			for _, s := range serials {
				s.Update()
			}
			time.Sleep(10 * time.Microsecond)
		}
	}()
	cl.Debug("-----------------------------------------------------")

	pq := queue.NewPriorityQueue(32, bus.NewEventClient("priorityqueue", topic.WavegenSend))
	pq.Event.Subscribe(topic.WavegenQueue)

	pool := packet.NewPacketPool(dcc.MaxPacketSize)

	w, err := InitWavegen(machine.GPIO22, machine.GPIO9, pool.DiscardPacket, bus)
	if err != nil {
		panic(err)
	}

	profiles := motor.ShieldEX8874
	// profiles := motor.ArduinoMotorShieldRev3
	motors := InitMotor(profiles, bus)

	tracks := make(map[string]*track.Track)
	for id := range motors {
		tracks[id] = track.NewTrack(id, track.TrackModeMain, bus.NewEventClient("track"+id, topic.MotorControl+id))
		tracks[id].Event.Subscribe(topic.MotorControl + id)
	}

	board := shared.BoardInfo{
		Board:      board,
		GitSHA:     gitSHA,
		ShieldName: shieldName,
		Version:    version,
	}
	d := dcc.NewDCC(board, w, pool, bus.NewEventClient("dcc", topic.BroadcastDex))
	cl.Debug("DCC initialized")

	// The wavegen and priority queue loops stall frequently so they get their own goroutines
	go w.Loop()
	go pq.Loop()

	go func() {
		for {
			dex.Update() // DCC-EX command processing
			time.Sleep(100 * time.Millisecond)
		}
	}()
	go func() {
		for {
			d.Update() // DCC main loop
			time.Sleep(5 * time.Millisecond)
		}
	}()

	for {
		// Handle track status events and power monitoring (short-circuit detection)
		for _, t := range tracks {
			t.Update()
		}
		for _, m := range motors {
			m.Update()
		}
		// runtime.Gosched()
		time.Sleep(100 * time.Microsecond) // FIXME: Cleanup - this should be a tight loop eventually
	}
}

func InitWavegen(signalPin, brakePin machine.Pin, packetReturn func(*packet.Packet), bus *event.EventBus) (*wavegen.Wavegen, error) {
	signalPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	brakePin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	config := wavegen.WavegenConfig{
		Mode:         wavegen.NormalMode,
		SignalPin:    signalPin,
		BrakePin:     brakePin,
		PacketReturn: packetReturn,
	}

	w, err := wavegen.NewWavegen(config, bus.NewEventClient("wavegen", topic.WavegenQueue))
	if err != nil {
		return nil, err
	}
	w.Event.Subscribe(topic.WavegenSend)
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
		motors[id] = motor.NewMotor(id, profile, bus.NewEventClient(id, topic.MotorStatus+id))
		motors[id].Event.Subscribe(topic.MotorControl + id)
	}

	return motors
}

type Serialer interface {
	WriteByte(c byte) error
	Write(data []byte) (n int, err error)
	Buffered() int
	ReadByte() (byte, error)
	Configure(config machine.UARTConfig) error
}
