# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

beacon-dcc is a TinyGo-based DCC (Digital Command Control) command station for Raspberry Pi RP2350 microcontrollers. It controls model train decoders using the NMRA DCC protocol and speaks the DCC-EX serial command protocol.

- **Language:** Go (TinyGo subset)
- **Target:** RP2350 (metro-rp2350 board)
- **Build constraint:** `//go:build rp` — main.go and serial.go only compile for RP hardware
- **Single external dependency:** `github.com/tinygo-org/pio` (PIO state machine bindings)

## Build Commands

```bash
make flash          # Build and flash to device
make build          # Build ELF binary
make uf2            # Build UF2 for USB mass storage flashing
make gdb            # Debug with GDB via OpenOCD
```

Environment variables: `TINYGO`, `GOTOOLCHAIN` (default: go1.25.8), `TARGET` (default: metro-rp2350), `SHIELD` (default: EX-MotorShield8874), `SERIAL` (default: usb).

## Testing

```bash
go test ./pkg/...                    # Run all tests
go test ./pkg/event                  # Run tests for a single package
go test ./pkg/packet -run TestName   # Run a specific test
```

Tests exist for: `pkg/event`, `pkg/packet`, `pkg/queue`, `pkg/serial`. Tests use standard `testing.T` with channels and `time.After` for timeout assertions. Note: packages under `//go:build rp` (main, wavegen, motor, track, adc) cannot be tested on host.

## Architecture

The system uses an **event bus** (`pkg/event`) for all inter-component communication. Components never call each other directly — they publish and subscribe to named topics.

```
Serial ──publish──> EventBus ──deliver──> DCCEX ──publish──> EventBus ──deliver──> DCC
                       │                                        │
                       ├──> Track (power/mode mgmt)             ├──> PriorityQueue
                       └──> Serial (output broadcast)           └──> Wavegen (PIO signal)
```

### Event Bus (`pkg/event`)

- **EventBus**: Central pub/sub router. Topic-based subscriptions with `map[topic]map[clientID]chan Event`.
- **EventClient**: Per-component wrapper providing `Publish()`, `PublishTo()`, `Subscribe()`, `Diag()`, `Debug()`. Created via `bus.NewEventClient(name, defaultTopic, optionalBufSize)`.
- **Non-blocking publish**: Events are dropped if a subscriber's channel is full (no backpressure).
- **Topics** defined as constants in `pkg/topic/topic.go`.

### Data Flow

1. **Serial input** → published to `rxcmd:serial`
2. **DCCEX** subscribes to `rxcmd:serial`, parses commands, calls registered handlers (e.g., `CmdThrottle`)
3. **DCC** generates packets → published to `wavegen:queue`
4. **PriorityQueue** reorders by priority → published to `wavegen:send`
5. **Wavegen** encodes packets for PIO state machine → DCC signal on track

### Key Packages

| Package | Purpose |
|---------|---------|
| `pkg/event` | EventBus and EventClient — all inter-component messaging |
| `pkg/topic` | Topic string constants |
| `pkg/dcc` | DCC protocol: loco state, speed/function packets, reminder loop |
| `pkg/dccex` | DCC-EX command parsing and handler dispatch |
| `pkg/packet` | Packet struct with pool (`sync.Pool`) and priority levels |
| `pkg/queue` | Min-heap priority queue with duplicate detection |
| `pkg/wavegen` | PIO-based DCC signal generation (modes: Normal, NoCutout, Service) |
| `pkg/motor` | Motor shield drivers with overcurrent detection via ADC |
| `pkg/track` | Track abstraction — mode (Main/Prog) and power management |
| `pkg/serial` | Serial I/O with event-based output; PseudoSerial wraps TinyGo default serial |
| `pkg/adc` | ADC wrapper for current sensing |
| `pkg/shared` | Hardware interfaces (Pin, StateMachine, PWM, ADC) and BoardInfo |

### Concurrency Model

- Each component has an `Update()` method called on a ticker via `RunEvery()` in its own goroutine
- Wavegen and PriorityQueue use blocking `Loop()` methods in dedicated goroutines
- Track updates run in the main goroutine (1ms interval)
- Loco state protected by `sync.RWMutex`; event bus subscriptions by `sync.Mutex`
- Packet pool uses `sync.Pool` to minimize GC pressure

### Initialization Pattern (main.go)

Components are wired in `main()`:
1. Create `EventBus`
2. Create each component with `bus.NewEventClient(...)`, specifying default publish topic
3. Call `Subscribe(...)` on each client for topics it consumes
4. Register command handlers with DCCEX (`RegisterCommandHandler`)
5. Launch goroutines for each component's loop

### Motor Shield Profiles

Hardware pin mappings and current limits are defined in `pkg/motor/profile.go`. Two profiles exist: `ShieldEX8874` and `ArduinoMotorShieldRev3`. Selected in `main()`.
