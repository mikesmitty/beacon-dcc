package topic

const (
	// BroadcastDex is the topic where replies to DCC-EX commands are published.
	BroadcastDex = "broadcast:dex"
	// BroadcastDebug is the topic where debug messages are published in the DCC-EX format.
	BroadcastDebug = "broadcast:debug"
	// BroadcastDiag is the topic where diagnostic messages are published.
	BroadcastDiag = "broadcast:diag"

	// ReceiveCmdSerial is the topic where commands received from serial input are published.
	ReceiveCmdSerial = "rxcmd:serial"

	TrackModeJoin   = "track:mode:join"
	TrackModeUnjoin = "track:mode:unjoin"
	TrackPowerOn    = "track:power:on"
	TrackPowerOff   = "track:power:off"
	TrackStatus     = "track:status"

	// WavegenQueue is the topic where packets are queued for the wave generator.
	WavegenQueue = "wavegen:queue"
	// WavegenSend is the topic where packets are sent to the wave generator for processing.
	// This should only be used for communication between the queue and the wave generator.
	WavegenSend = "wavegen:send"
)
