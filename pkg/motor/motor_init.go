//go:build rp

package motor

import (
	"machine"
	"time"

	"github.com/mikesmitty/beacon-dcc/pkg/shared"
)

func (m *Motor) Init(trackId string) error {
	m.trackId = trackId

	if m.PowerPin != shared.NoPin {
		m.PowerPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	}
	if m.FaultPin != shared.NoPin {
		m.FaultPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	}

	// Zero out the current reading with the power off
	m.setPowerMode(PowerModeOff)
	m.ADC.InitADC()
	time.Sleep(1 * time.Second) // Let ADC stabilize
	m.ADC.SetBaseline()
	return nil
}
