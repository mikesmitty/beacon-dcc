//go:build rp

package main

import (
	"machine"
	"strings"
	"time"

	"github.com/mikesmitty/beacon-dcc/pkg/adc"
	"github.com/mikesmitty/beacon-dcc/pkg/dcc"
	"github.com/mikesmitty/beacon-dcc/pkg/dccex"
	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/motor"
	"github.com/mikesmitty/beacon-dcc/pkg/packet"
	"github.com/mikesmitty/beacon-dcc/pkg/queue"
	"github.com/mikesmitty/beacon-dcc/pkg/shared"
	"github.com/mikesmitty/beacon-dcc/pkg/topic"
	"github.com/mikesmitty/beacon-dcc/pkg/track"
	"github.com/mikesmitty/beacon-dcc/pkg/wavegen"
)

var (
	board      string
	gitSHA     string
	shieldName string
	version    string
)

func main() {
	println("main.main has launched") // FIXME: Cleanup

	boardInfo := shared.BoardInfo{
		Board:      board,
		GitSHA:     gitSHA,
		ShieldName: shieldName,
		Version:    strings.TrimPrefix(version, "v"),
	}

	bus := event.NewEventBus()

	serials := InitSerials(bus, topic.ReceiveCmdSerial, topic.BroadcastDex, topic.BroadcastDiag, topic.BroadcastDebug)
	go RunEvery(serials.Update, 50*time.Millisecond)

	dex := dccex.NewDCCEX(boardInfo, bus.NewEventClient("dccex", topic.BroadcastDex))
	dex.Event.Subscribe(topic.ReceiveCmdSerial, topic.TrackStatus)

	pq := queue.NewPriorityQueue(32, bus.NewEventClient("priorityqueue", topic.WavegenSend))
	pq.Event.Subscribe(topic.WavegenQueue)

	pool := packet.NewPacketPool(dcc.MaxPacketSize)

	w, err := InitWavegen(machine.GPIO22, machine.GPIO9, pool.DiscardPacket, bus)
	if err != nil {
		panic(err)
	}

	profiles := motor.ShieldEX8874
	// profiles := motor.ArduinoMotorShieldRev3
	tracks, _, err := InitTracks(profiles, bus)
	if err != nil {
		panic(err)
	}

	d := dcc.NewDCC(boardInfo, w, pool, bus.NewEventClient("dcc", topic.BroadcastDex))

	// The wavegen and priority queue loops stall frequently so they get their own goroutines
	go w.Loop()
	go pq.Loop()

	go RunEvery(dex.Update, 50*time.Millisecond) // DCC-EX command processing
	go RunEvery(d.Update, 5*time.Millisecond)    // DCC main loop

	for {
		// Handle track status events and power monitoring (short-circuit detection)
		for _, t := range tracks {
			t.Update()
		}
		time.Sleep(1 * time.Millisecond)
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

func InitTracks(profiles map[string]motor.MotorShieldProfile, bus *event.EventBus) (map[string]*track.Track, map[string]*motor.Motor, error) {
	adcs := map[string]motor.ADC{
		"A": adc.NewADC(machine.ADC1),
		"B": adc.NewADC(machine.ADC2),
	}

	motors := make(map[string]*motor.Motor)
	for id, profile := range profiles {
		profile.ADC = adcs[id]
		motors[id] = motor.NewMotor(profile)
	}

	tracks := make(map[string]*track.Track)
	for id := range motors {
		t, err := track.NewTrack(id, track.TrackModeMain, motors[id], bus.NewEventClient("track"+id, topic.TrackStatus))
		if err != nil {
			return nil, nil, err
		}
		tracks[id] = t
	}

	return tracks, motors, nil
}

func RunEvery(fn func(), interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			fn()
		}
	}
}
