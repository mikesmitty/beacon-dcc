package dccex

import "fmt"

func BroadcastForgetLoco(loco uint16) string {
	// Emergency stop (forward)
	return fmt.Sprintf("<l %d 0 129 0>\n<- %d>\n", loco, loco)
}

func BroadcastStopLoco(loco uint16, forward bool) string {
	var speed uint8
	if forward {
		speed = 128
	}
	return BroadcastLocoState(loco, speed, 0)
}

func BroadcastLocoState(loco uint16, speedStep uint8, functions uint32) string {
	return fmt.Sprintf("<l %d 0 %d %d>\n", loco, speedStep, functions)
}
