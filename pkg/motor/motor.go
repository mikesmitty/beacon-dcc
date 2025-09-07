package motor

import (
	"errors"
	"time"

	"github.com/mikesmitty/beacon-dcc/pkg/event"
	"github.com/mikesmitty/beacon-dcc/pkg/pwm"
	"github.com/mikesmitty/beacon-dcc/pkg/shared"
	"github.com/mikesmitty/beacon-dcc/pkg/topic"
	"github.com/mikesmitty/beacon-dcc/pkg/track"
)

const (
	// Base for wait time until power is turned on again
	minOvercurrentBackoff = 40 * time.Millisecond
	// Time after we consider all faults old and forgotten
	overcurrentFaultWindow = 5 * time.Second
	// Time after which we consider a ALERT over
	alertWindow = 20 * time.Millisecond
	// How long to ignore fault pin if current is under limit
	ignoreFaultNormalTimeout = 100 * time.Millisecond
	// How long to ignore fault pin if current is higher than limit
	ignoreFaultOverCurrentTimeout = 5 * time.Millisecond
	// How long to wait between overcurrent and turning off
	overcurrentTimeout = 100 * time.Millisecond
	// Upper limit for retry period
	retryWindow = 10 * time.Second
)

type MotorShieldProfile struct {
	PowerPin    shared.Pin
	InvertPower bool // True if the power PWM pin is inverted (active low)
	SignalPin   shared.Pin
	BrakePin    shared.Pin
	ADC         ADC
	SenseFactor float32 // mA per ADC unit
	MaxCurrent  uint16  // Max current in mA
	FaultPin    shared.Pin
}

type Motor struct {
	MotorShieldProfile
	brakePWM *pwm.SimplePWM
	pwrPWM   *pwm.SimplePWM
	trackId  string

	currentLimit     uint16 // ADC-equivalent reading
	lastReading      uint16 // Last ADC reading
	progCurrentLimit uint16 // ADC-equivalent reading
	progMode         bool   // True if this is a programming track
	zeroReading      uint16 // Zero reading for the ADC

	lastBadAlertCheck  time.Time
	lastStateTime      map[PowerMode]time.Time
	overcurrentBackoff time.Duration
	prevState          PowerMode
	state              PowerMode

	Event *event.EventClient
}

var _ track.Driver = (*Motor)(nil)

// NewMotor initializes a new Motor instance with the given ADC pins and current sense factor.
// Current sense factor is milliamps of output current per ADC unit.
func NewMotor(profile MotorShieldProfile) *Motor {
	// DCC-EX profiles use the native 12-bit ADC range of most MCUs, but TinyGo normalizes to 16-bit
	profile.SenseFactor = profile.SenseFactor * 16

	m := &Motor{
		MotorShieldProfile: profile,

		currentLimit:     uint16(float32(profile.MaxCurrent) * profile.SenseFactor),
		progCurrentLimit: uint16(250.0 * profile.SenseFactor), // 250mA

		lastStateTime: make(map[PowerMode]time.Time),
		prevState:     PowerModeNone,
		state:         PowerModeOff,
	}

	return m
}

func (m *Motor) SetEventClient(cl *event.EventClient) {
	m.Event = cl
}

func (m *Motor) Init(trackId string) error {
	m.trackId = trackId
	// Zero out the current reading with the power off
	m.setPowerMode(PowerModeOff)
	m.ADC.InitADC()
	time.Sleep(1 * time.Second) // Let ADC stabilize
	m.ADC.SetBaseline()
	return nil
}

func (m *Motor) Power() (bool, error) {
	switch m.state {
	case PowerModeOff:
		return false, nil
	case PowerModeOn:
		return true, nil
	case PowerModeAlert:
		return true, errors.New("Motor is in alert state")
	case PowerModeOverload:
		return false, errors.New("Motor is in overload state")
	default:
		return false, nil
	}
}

func (m *Motor) SetPower(on bool) error {
	if on {
		m.setPowerMode(PowerModeOn)
	} else {
		m.setPowerMode(PowerModeOff)
	}
	return nil
}

// Check for shorts or ack responses on programming tracks
func (m *Motor) Update() {
	/* FIXME: Cleanup
	select {
	case evt := <-m.Event.Receive:
		switch msg := evt.Data.(type) {
		case string:
			m.handleStringEvent(msg)
		default:
			m.Event.Debug("Received unknown event type: %T", evt.Data)
		}
	default:
		// No event to process
	}
	*/

	// TODO: Add prog-track config?

	switch m.state {
	case PowerModeOff:
		m.overcurrentBackoff = minOvercurrentBackoff

	case PowerModeOn:
		m.handleOnState()

	case PowerModeAlert:
		firstTime := m.prevState != PowerModeAlert
		m.handleAlertState(firstTime)

	case PowerModeOverload:
		m.handleOverloadState()
	}

	if m.prevState != m.state {
		var on bool
		switch m.state {
		case PowerModeOn, PowerModeAlert:
			on = true
		}
		m.Event.PublishTo(topic.TrackStatus, track.TrackStatus{
			ID:       m.trackId,
			Power:    on,
			Alert:    m.state == PowerModeAlert,
			Overload: m.state == PowerModeOverload,
		})
	}
	m.prevState = m.state
}

// FIXME: Cleanup
func (m *Motor) handleStringEvent(msg string) {
	newState := m.state
	switch msg {
	case "on":
		newState = PowerModeOn
	case "off":
		newState = PowerModeOff
	default:
		m.Event.Debug("Invalid motor state: %s", msg)
		return
	}

	m.setPowerMode(newState)
}

func (m *Motor) handleOnState() {
	// FIXME: Cleanup
	// cF = checkFault() // returns negative if the ADC isn't initialized?
	fault := m.checkFault()
	overcurrent, mA := m.checkOverCurrent()

	if !overcurrent && !fault {
		if time.Since(m.lastStateTime[PowerModeOn]) > overcurrentFaultWindow {
			m.overcurrentBackoff = minOvercurrentBackoff
		}
		return
	}

	if overcurrent && fault {
		m.Event.Diag("TRACK %s ALERT FAULT %dmA", m.trackId, mA)
	} else if overcurrent {
		m.Event.Debug("TRACK %s ALERT %dmA", m.trackId, mA)
	} else {
		m.Event.Debug("TRACK %s FAULT", m.trackId)
	}

	m.setPowerMode(PowerModeAlert)

	// FIXME: Implement?
	// if ((trackMode & TRACK_MODIFIER_AUTO) && (trackMode & (TRACK_MODE_MAIN|TRACK_MODE_EXT|TRACK_MODE_BOOST))) {
	// 	DIAG(F("TRACK %c INVERT"), trackno + 'A');
	// 	invertOutput();
	// }
}

func (m *Motor) handleAlertState(firstTime bool) {
	timeSinceAlert := time.Since(m.lastStateTime[PowerModeAlert])

	fault := m.checkFault()
	overcurrent, mA := m.checkOverCurrent()

	if fault {
		m.limitInrush(true)
		m.lastBadAlertCheck = time.Now()

		timeout := ignoreFaultNormalTimeout
		if overcurrent {
			timeout = ignoreFaultOverCurrentTimeout
		}

		if timeSinceAlert < timeout {
			return
		}

		if firstTime {
			m.Event.Diag("TRACK %s FAULT PIN (%s ignore)", m.trackId, timeout)
			return
		}
		m.Event.Diag("TRACK %s FAULT PIN detected after %s. Pause %s", m.trackId, timeSinceAlert, overcurrentTimeout)
		m.limitInrush(false)
		m.setPowerMode(PowerModeOverload)
		return
	}

	if overcurrent {
		m.lastBadAlertCheck = time.Now()

		if timeSinceAlert < overcurrentTimeout {
			return
		}

		if firstTime {
			m.Event.Diag("TRACK %s CURRENT (%s ignore) %dmA", m.trackId, overcurrentTimeout, mA)
			return
		}
		m.Event.Diag("TRACK %s POWER OVERLOAD %dmA (max %dmA) detected after %s. Pause %s", m.trackId, mA, m.MaxCurrent, timeSinceAlert, overcurrentTimeout)
		m.limitInrush(false)
		m.setPowerMode(PowerModeOverload)
		return
	}

	// If we reach here, there are no faults or overcurrent conditions
	if timeSinceAlert > alertWindow {
		m.Event.Diag("TRACK %s NORMAL (after %s/%s) %dmA", m.trackId, timeSinceAlert, alertWindow, mA)
		m.limitInrush(false)
		m.setPowerMode(PowerModeOn)
	} else if timeSinceAlert > alertWindow/2 {
		m.limitInrush(false)
	}
}

func (m *Motor) handleOverloadState() {
	timeSinceOverload := time.Since(m.lastStateTime[PowerModeOverload])
	if timeSinceOverload < m.overcurrentBackoff {
		return
	}

	// Exponential backoff for retrying the power on
	m.overcurrentBackoff *= 2
	if m.overcurrentBackoff > retryWindow {
		m.overcurrentBackoff = retryWindow
	}

	/* FIXME: Implement?
	#ifdef EXRAIL_ACTIVE
		DIAG(F("Calling EXRAIL"));
	    RMFT2::powerEvent(trackno, true); // Tell EXRAIL we have an overload
	#endif
	*/

	m.Event.Diag("TRACK %s POWER RESTORE (after %s)", m.trackId, timeSinceOverload)
	m.setPowerMode(PowerModeAlert)
}

func (m *Motor) checkFault() bool {
	if m.FaultPin == shared.NoPin {
		return false
	}
	return m.FaultPin.Get()
}

func (m *Motor) checkOverCurrent() (bool, uint16) {
	// Refresh the current reading
	m.lastReading = m.ADC.Get()

	mA := uint16(float32(m.lastReading) / m.SenseFactor)
	if m.lastReading > m.currentLimit || (m.progMode && m.lastReading > m.progCurrentLimit) {
		m.Event.Debug("TRACK ALERT %dmA", mA)
		// FIXME: Cleanup
		// DIAG(F("TRACK %c ALERT %s %dmA"), trackno + 'A', cF ? "FAULT" : "", mA);
		return true, mA
	}
	return false, mA
}

func (m *Motor) setPowerMode(mode PowerMode) {
	if m.state == mode {
		return
	}
	m.Event.Debug("Motor %s setPowerMode %s", m.trackId, mode)

	m.lastStateTime[mode] = time.Now()

	if mode == PowerModeOn || mode == PowerModeAlert {
		// When switching a track On, we need to check the current offset with the pin OFF
		// to set the zero value for the ADC (in case of a driver that centers zero in the
		// middle of the range and go up/down based on direction)
		if m.state == PowerModeOff && m.ADC != nil {
			m.zeroReading = m.ADC.Get()
			m.Event.Diag("Motor %s zeroReading=%d", m.trackId, m.zeroReading)
		}

		m.PowerPin.Set(!m.InvertPower) // High

		if m.progMode {
			// FIXME: Implement
			// DCCWaveform::progTrack.clearResets();
		}
	} else {
		m.PowerPin.Set(m.InvertPower) // Low
	}

	m.state = mode
}

func (m *Motor) limitInrush(on bool) {
	if m.BrakePin == shared.NoPin {
		return
	}
	// FIXME: Implement?
	// if ( !(trackMode & (TRACK_MODE_MAIN | TRACK_MODE_PROG | TRACK_MODE_EXT | TRACK_MODE_BOOST)))
	//   return;

	// FIXME: Will PWM and PIO brake operation work on the same pin?

	if on {
		if m.brakePWM == nil {
			brakePwm, err := pwm.NewPWM(m.BrakePin, pwm.MaxFreq, 80.0)
			if err != nil {
				m.Event.Diag("Failed to initialize PWM for brake pin: %v", err)
				return
			}
			m.brakePWM = brakePwm
		}
	} else if m.brakePWM != nil {
		m.brakePWM.SetDuty(0.0)
		m.brakePWM = nil
	}
}

type ADC interface {
	InitADC()
	Get() uint16
	SetBaseline()
}
