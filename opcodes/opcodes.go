// Package opcodes defines the shared opcode constants for the Lattice mesh protocol.
// C headers in c/ are generated from these constants — run "go generate ./..." to regenerate.
//
//go:generate go run ../cmd/gen-headers/main.go
package opcodes

// Serial command opcodes — byte 0 of the data payload in MessageTypeSerialCmdBroadcast frames.
const (
	// Health reporting — bidirectional between server and nodes.
	OpHealthReq    = byte(0xB0) // Server → node: request health report; payload: [B0] (no body)
	OpHealthReport = byte(0xB1) // Node (serial) → server: health status; payload: [B1][1B adapterType][6B mac][4B uptimeSec LE]
	OpNodeHealth   = byte(0xB2) // Node (non-serial) → server via serial adapter; payload: [B2][1B adapterType][6B mac][4B uptimeSec LE]

	// Node → server: route reporting
	OpRouteReport  = byte(0xB3) // Node→server: routing path; payload: [B3][1B path_len][path_len × 6B MACs]

	// Server → node: management
	OpNodeIdSet  = byte(0xC0) // Server → node: assign logical node ID
	OpConfigSet  = byte(0xC1) // Server → node: set adapter type and config; payload: [C1][6B targetMac][1B adapterType]
	OpTxPowerSet = byte(0xC2) // Server → node: set TX power preset; payload: [C2][1B preset: 0=short 1=indoor 2=outdoor]

	// Output adapter commands (server → output node)
	OpLEDSolid = byte(0xD0) // Set LED strip to solid colour; payload: [D0][1B r][1B g][1B b]
	OpLEDOff   = byte(0xD1) // Turn LED strip off; payload: [D1] (no body)
	OpLEDBlink = byte(0xD2) // Blink LED; payload: [D2][1B r][1B g][1B b][1B interval_hi][1B interval_lo]
	OpRelaySet = byte(0xD8) // Set relay state; payload: [D8][1B: 0x00=off 0x01=on]

	// Input adapter events (node → server)
	OpCommandAck = byte(0xE0) // Node → server: acknowledge a received command
)
