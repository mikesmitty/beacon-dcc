package dccex

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
)

func (d *DCCEX) cmdStatus(resp *bytes.Buffer, cmd byte, params [][]byte) error {
	resp.Write(fmt.Appendf(nil, "<iDCC-EX V-%s / %s / %s G-%s>\n", d.info.Version, d.info.Board, d.info.ShieldName, d.info.GitSHA))

	for _, k := range slices.Sorted(maps.Keys(d.tracks)) {
		v := d.tracks[k]
		var power byte
		if v.Power {
			power = 1
		}
		resp.Write(fmt.Appendf(nil, "<p%d %s>\n", power, k))
	}

	/* FIXME: Implement
	   CommandDistributor::broadcastPower(); // <s> is the only "get power status" command we have
	   Turnout::printAll(stream); //send all Turnout states
	   Sensor::printAll(stream);  //send all Sensor  states
	*/
	return nil
}

/* FIXME: Cleanup
bool TrackManager::getPower(byte t, char s[]) {
  if (track[t]) {
    s[0] = track[t]->getPower() == POWERMODE::ON ? '1' : '0';
    s[2] = t + 'A';
    return true;
  }
}

void  CommandDistributor::broadcastPower() {
  char pstr[] = "? x";
  for(byte t=0; t<TrackManager::MAX_TRACKS; t++)
    if (TrackManager::getPower(t, pstr))
	// "<p1 A>\n"
	// "<p0 B>\n"
      broadcastReply(COMMAND_TYPE, F("<p%s>\n"),pstr);

  // FIXME: Start again from here
  byte trackcount=0;
  byte oncount=0;
  byte offcount=0;
  for(byte t=0; t<TrackManager::MAX_TRACKS; t++) {
    if (TrackManager::isActive(t)) {
      trackcount++;
      // do not call getPower(t) unless isActive(t)!
      if (TrackManager::getPower(t) == POWERMODE::ON)
	oncount++;
      else
	offcount++;
    }
  }
  //DIAG(F("t=%d on=%d off=%d"), trackcount, oncount, offcount);

  char state='2';
  if (oncount==0 || offcount == trackcount)
    state = '0';
  else if (oncount == trackcount) {
    state = '1';
  }

  if (state != '2')
    broadcastReply(COMMAND_TYPE, F("<p%c>\n"),state);

  // additional info about MAIN, PROG and JOIN
  bool main=TrackManager::getMainPower()==POWERMODE::ON;
  bool prog=TrackManager::getProgPower()==POWERMODE::ON;
  bool join=TrackManager::isJoined();
  //DIAG(F("m=%d p=%d j=%d"), main, prog, join);
  const FSH * reason=F("");
  if (join) {
    reason = F(" JOIN"); // with space at start so we can append without space
    broadcastReply(COMMAND_TYPE, F("<p1%S>\n"),reason);
  } else {
    if (main) {
      //reason = F("MAIN");
      broadcastReply(COMMAND_TYPE, F("<p1 MAIN>\n"));
    }
    if (prog) {
      //reason = F("PROG");
      broadcastReply(COMMAND_TYPE, F("<p1 PROG>\n"));
    }
  }
#ifdef CD_HANDLE_RING
  // send '1' if all main are on, otherwise global state (which in that case is '0' or '2')
  broadcastReply(WITHROTTLE_TYPE, F("PPA%c\n"), main?'1': state);
#endif

  LCD(2,F("Power %S%S"),state=='1'?F("On"): ( state=='0'? F("Off") : F("SC") ),reason);
}
*/
