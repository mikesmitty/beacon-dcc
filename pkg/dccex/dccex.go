package dccex

import (
	"bytes"
	"fmt"

	"github.com/mikesmitty/beacon-dcc/pkg/event"
)

/* DCC-EX Command Set
https://dcc-ex.com/reference/software/command-summary-consolidated.html

Character, Usage
/, |EX-R| interactive commands
-, Remove from reminder table
=, |TM| configuration
!, Emergency stop
@, Reserved for future use - LCD messages to JMRI
#, Request number of supported cabs/locos; heartbeat
+, WiFi AT commands
?, Reserved for future use
0, Track power off
1, Track power on
a, DCC accessory control
A, DCC extended accessory control
b, Write CV bit on main
B, Write CV bit
c, Request current command
C, configure the CS
d,
D, Diagnostic commands
e, Erase EEPROM
E, Store configuration in EEPROM
f, Loco decoder function control (deprecated)
F, Loco decoder function control
g,
G,
h,
H, Turnout state broadcast
i, Server details string
I, Turntable object command, control, and broadcast
j, Throttle responses
J, Throttle queries
k, Block exit  (Railcom)
K, Block enter (Railcom)
l, Loco speedbyte/function map broadcast
L, Reserved for LCC interface (implemented in EXRAIL)
m, message to throttles (broadcast output)
m, set momentum
M, Write DCC packet
n, Reserved for SensorCam
N, Reserved for Sensorcam
o, Neopixel driver (see also IO_NeoPixel.h)
O, Output broadcast
p, Broadcast power state
P, Write DCC packet
q, Sensor deactivated
Q, Sensor activated
r, Broadcast address read on programming track
R, Read CVs
s, Display status
S, Sensor configuration
t, Cab/loco update command
T, Turnout configuration/control
u, Reserved for user commands
U, Reserved for user commands
v,
V, Verify CVs
w, Write CV on main
W, Write CV
x,
X, Invalid command response
y,
Y, Output broadcast
z, Direct output
Z, Output configuration/control
*/

const (
	MaxParameters = 10
)

type HandlerFunc func(resp *bytes.Buffer, cmd byte, params [][]byte) error

type DCCEX struct {
	handlers map[byte]HandlerFunc

	Event *event.EventClient
}

func NewDCCEX(cl *event.EventClient) *DCCEX {
	d := &DCCEX{
		Event: cl,

		handlers: make(map[byte]HandlerFunc),
	}

	// Register built-in command handlers
	d.RegisterCommandHandler(cmdOn, '1')
	d.RegisterCommandHandler(cmdOff, '0')

	return d
}

func (d *DCCEX) Update() {
	select {
	case evt := <-d.Event.Receive:
		msg, ok := evt.Data.(*bytes.Buffer)
		if !ok {
			return
		}

		input := msg.Bytes()
		if len(input) == 0 {
			return
		}

		response, err := d.handleCommand(evt.ClientID, input)
		if response.Len() == 0 && err == nil {
			// Most likely another module will respond
			return
		}
		if err != nil {
			response.WriteString("<X>")
			// FIXME: Log error
		}
		d.Event.Publish(response)
	default:
		// No event to process
	}
}

// FIXME: Is there a use for clientId here?
func (d *DCCEX) handleCommand(clientId string, input []byte) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	if len(input) == 0 {
		return buf, fmt.Errorf("Empty command received")
	}

	words := bytes.Split(input, []byte{' '})
	cmd := input[0]
	params := words[1:]
	if len(params) > MaxParameters {
		return buf, fmt.Errorf("Too many parameters: %d", len(params))
	}

	if handler, ok := d.handlers[cmd]; ok {
		err := handler(buf, cmd, params)
		if err != nil {
			return buf, fmt.Errorf("Error handling command %c: %v", cmd, err)
		}
		return buf, nil
	}

	return buf, fmt.Errorf("Unsupported command: %c", cmd)
}

func (d *DCCEX) RegisterCommandHandler(handler HandlerFunc, opCodes ...byte) error {
	for _, opCode := range opCodes {
		if _, exists := d.handlers[opCode]; exists {
			return fmt.Errorf("Command handler already registered for: %c", opCode)
		}
		d.handlers[opCode] = handler
	}
	return nil
}
