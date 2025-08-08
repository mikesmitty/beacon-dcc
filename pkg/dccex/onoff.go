package dccex

import (
	"bytes"
	"fmt"
)

/*
<onOff [track]> - Turn power on or off to the MAIN and PROG tracks

https://dcc-ex.com/reference/software/command-summary-consolidated.html#onoff-track-turn-power-on-or-off-to-the-main-and-prog-tracks
Also allows joining the MAIN and PROG tracks together.

Parameters:

	> onOff: one of
	   • 1 = on
	   • 0 = off
	> track: one of
	   • blank = Both Main and Programming Tracks
	   • MAIN = Main track
	   • PROG = Programming Track
	   • JOIN = Join the Main and Programming tracks temporarily

Note: While <1 JOIN> is valid, <0 JOIN> is not.

Response:

The following is not a direct response, but rather a broadcast that will be triggered as a result of any power state changes.

<pOnOFF [track]>

Notes:

The use of the JOIN function allows the DCC signal for the MAIN track to also be sent to the PROG track. This allows the prog track to act as a siding (or similar) in the main layout even though it is isolated electrically and connected to the programming track output.

However, it is important that the prog track wiring be in the same phase as the main track i.e. when the left rail is high on MAIN, it is also high on PROG. You may have to swap the wires to your prog track to make this work.

If you drive onto a programming track that is “joined” and enter a programming command, the track will automatically switch to a programming track. If you use a compatible Throttle, you can then send the join command again and drive off the track onto the rest of your layout!

In some split Motor Shield hardware configurations JOIN will not be able to work.

While <1 JOIN> is valid, <0 JOIN> is not.

Examples:

	all tracks off: <0>
	all tracks on <1>
	join: <1 JOIN>

Example Responses:

	all tracks off: <p0>
	all tracks on <p1>
	join: <p1 JOIN>
*/
func cmdOn(resp *bytes.Buffer, cmd byte, params [][]byte) error {
	for i := range params {
		params[i] = bytes.ToUpper(params[i])
	}

	switch len(params) {
	case 0:
		// All tracks
		// FIXME: Implement
		// TrackManager::setTrackPower(TRACK_ALL, POWERMODE::ON);

		// resp.WriteString("<p1>") // FIXME: Done by TrackManager?
	case 1:
		if bytes.Equal(params[0], []byte("MAIN")) { // <1 MAIN>
			// FIXME: Implement
			// TrackManager::setTrackPower(TRACK_MODE_MAIN, POWERMODE::ON);
			// #ifndef DISABLE_PROG
		} else if bytes.Equal(params[0], []byte("JOIN")) { // <1 JOIN>
			// else if (p[0] == "JOIN"_hk) {  // <1 JOIN>
			// FIXME: Implement
			// TrackManager::setJoin(true);
			// TrackManager::setTrackPower(TRACK_MODE_MAIN|TRACK_MODE_PROG, POWERMODE::ON);
		} else if bytes.Equal(params[0], []byte("PROG")) { // <1 PROG>
			// FIXME: Implement
			// TrackManager::setJoin(false);
			// TrackManager::setTrackPower(TRACK_MODE_PROG, POWERMODE::ON);
			// #endif
		} else if len(params[0]) == 1 && params[0][0] >= 'A' && params[0][0] <= 'H' { // <1 A-H>
			trackNum := params[0][0] - 'A'
			_ = trackNum // FIXME: Cleanup
			// FIXME: Implement
			// TrackManager::setTrackPower(POWERMODE::ON, t);
		} else {
			return fmt.Errorf("Unsupported parameter: %s", params[0])
		}
	default:
		return fmt.Errorf("Too many parameters")
	}
	return nil
}

func cmdOff(resp *bytes.Buffer, cmd byte, params [][]byte) error {
	for i := range params {
		params[i] = bytes.ToUpper(params[i])
	}

	switch len(params) {
	case 0:
		// All tracks
		// FIXME: Implement
		// TrackManager::setJoin(false);
		// TrackManager::setTrackPower(TRACK_ALL, POWERMODE::OFF);

		// resp.WriteString("<p0>") // FIXME: Done by TrackManager?
	case 1:
		if bytes.Equal(params[0], []byte("MAIN")) { // <0 MAIN>
			// FIXME: Implement
			// TrackManager::setJoin(false);
			// TrackManager::setTrackPower(TRACK_MODE_MAIN, POWERMODE::OFF);
		} else if bytes.Equal(params[0], []byte("PROG")) { // <0 PROG>
			// FIXME: Implement
			// TrackManager::setJoin(false);
			// TrackManager::progTrackBoosted=false;  // Prog track boost mode will not outlive prog track off
			// TrackManager::setTrackPower(TRACK_MODE_PROG, POWERMODE::OFF);
		} else if len(params[0]) == 1 && params[0][0] >= 'A' && params[0][0] <= 'H' { // <0 A-H>
			trackNum := params[0][0] - 'A'
			_ = trackNum // FIXME: Cleanup
			// FIXME: Implement
			// TrackManager::setJoin(false);
			// TrackManager::setTrackPower(POWERMODE::OFF, t);
			// //StringFormatter::send(stream, F("<p0 %c>\n"), trackNum+'A');
		} else {
			return fmt.Errorf("Unsupported parameter: %s", params[0])
		}
	default:
		return fmt.Errorf("Too many parameters")
	}
	return nil
}
