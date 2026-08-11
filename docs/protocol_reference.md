# Lattice mesh protocol reference

This is a consolidated reference for the Lattice mesh wire protocol: message types, opcodes,
adapter types, and the `MeshMessage` wire layout. Every value below was transcribed directly from
the Go source (the source of truth) and cross-checked against the generated C header — not copied
from prose elsewhere. If this doc and the source ever disagree, the source wins; re-derive from:

- `message/types.go` — message-type constants
- `opcodes/opcodes.go` — opcode constants
- `adapter/types.go` — adapter-type constants
- `message/message.go` — the `MeshMessage` struct and `WireSize`
- `c/mesh_message.h` — generated C struct and `static_assert` on total size

C headers are generated from the Go constants via `go generate ./...` (see `cmd/gen-headers`); the
Go files are always the ones to edit.

Verified against the `v0.6.0` release tag (commit `99cd30c`).

## Message types

Source: `message/types.go`. 6 constants, all `uint8`.

| Name | Value | Meaning |
| --- | --- | --- |
| `MeshTypeAdapterData` | 0 | Node data relayed toward server |
| `MeshTypeMasterBeacon` | 1 | Master→mesh: topology beacon |
| `MeshTypeEnrollment` | 2 | Node→master: enrollment request |
| `MeshTypeSerialCmdBroadcast` | 3 | Server→node: serial command broadcast |
| `MeshTypeJoinAck` | 4 | Server→node: enrollment approved |
| `MeshTypeRouteReport` | 5 | Node→server: routing path report |

## Opcodes

Source: `opcodes/opcodes.go`. Opcodes are byte 0 of the data payload carried in
`MeshTypeSerialCmdBroadcast` frames (the `Data` field of `MeshMessage`). **12 constants, all
`byte`, in 5 groups** (grouping and comments below are transcribed directly from the source; note
this is 12 opcodes, not 11 — verify against the source file directly if any other document states
a different count).

### Health reporting (bidirectional between server and nodes)

| Name | Hex | Meaning |
| --- | --- | --- |
| `OpHealthReq` | `0xB0` | Server → node: request health report; payload: `[B0]` (no body) |
| `OpHealthReport` | `0xB1` | Node (serial) → server: health status; payload: `[B1][1B adapterType][6B mac][4B uptimeSec LE]` |
| `OpNodeHealth` | `0xB2` | Node (non-serial) → server via serial adapter; payload: `[B2][1B adapterType][6B mac][4B uptimeSec LE]` |

### Node → server: route reporting

| Name | Hex | Meaning |
| --- | --- | --- |
| `OpRouteReport` | `0xB3` | Node→server: routing path; payload: `[B3][1B path_len][path_len × 6B MACs]` |

### Server → node: management

| Name | Hex | Meaning |
| --- | --- | --- |
| `OpNodeIdSet` | `0xC0` | Server → node: assign logical node ID |
| `OpConfigSet` | `0xC1` | Server → node: set adapter type and config; payload: `[C1][6B targetMac][1B adapterType]` |
| `OpTxPowerSet` | `0xC2` | Server → node: set TX power preset; payload: `[C2][1B preset: 0=short 1=indoor 2=outdoor]` |

### Output adapter commands (server → output node)

| Name | Hex | Meaning |
| --- | --- | --- |
| `OpLEDSolid` | `0xD0` | Set LED strip to solid colour; payload: `[D0][1B r][1B g][1B b]` |
| `OpLEDOff` | `0xD1` | Turn LED strip off; payload: `[D1]` (no body) |
| `OpLEDBlink` | `0xD2` | Blink LED; payload: `[D2][1B r][1B g][1B b][1B interval_hi][1B interval_lo]` |
| `OpRelaySet` | `0xD8` | Set relay state; payload: `[D8][1B: 0x00=off 0x01=on]` |

### Input adapter events (node → server)

| Name | Hex | Meaning |
| --- | --- | --- |
| `OpCommandAck` | `0xE0` | Node → server: acknowledge a received command |

## Adapter types

Source: `adapter/types.go`. 5 constants, all `int32`.

| Name | Value | Meaning |
| --- | --- | --- |
| `TypeUnknown` | 0 | Not yet configured |
| `TypeSerial` | 1 | Serial management (internal) |
| `TypePIR` | 2 | Passive infrared motion sensor (INPUT) |
| `TypeLED` | 3 | LED strip (OUTPUT) |
| `TypeRelay` | 4 | Relay switch (OUTPUT) |

Two helper functions in the same file classify these: `IsInput(t)` returns `true` only for
`TypePIR`; `IsOutput(t)` returns `true` for `TypeLED` and `TypeRelay` (`TypeUnknown` and
`TypeSerial` are neither).

## Wire layout

**Protocol version: v5** (per the `MeshMessage` doc comment in `message/message.go`).
**Total wire size: 200 bytes** — `message.WireSize = 200`, enforced at compile time by
`static_assert(sizeof(mesh_message) == 200, "mesh_message size changed — update server proto")`
in `c/mesh_message.h`.

The C struct is declared `__attribute__((packed))` (no compiler alignment padding), and the Go
struct's doc comment states "Field order matches the packed C struct — do not reorder without
updating the static_assert." Both files list the fields in the same order with the same sizes.
Byte offsets below are therefore fully determinable: each field starts immediately after the
previous one ends, starting at offset 0. The running total below reaches exactly 200, matching
`WireSize` and the `static_assert`.

| Offset | Size | Field (Go / C) | Type (Go / C) | `proto` tag | Purpose |
| --- | --- | --- | --- | --- | --- |
| 0 | 1 | `ProtoVersion` / `proto_version` | `uint8` / `uint8_t` | `10,uint32` | Protocol version carried by this frame (current: v5) |
| 1 | 1 | `MessageType` / `message_type` | `uint8` / `uint8_t` | `1,uint32` | Message-type discriminator; one of the `MeshType*` constants (see Message types) |
| 2 | 4 | `DataType` / `data_type` | `int32` / `int32_t` | `2,sint32` | Payload data-type discriminator (signed); semantics depend on `MessageType` |
| 6 | 6 | `OriginMacAddress` / `origin_mac_address` | `[6]byte` / `uint8_t[6]` | `3,bytes` | MAC address of the node that originated the frame |
| 12 | 6 | `TargetMacAddress` / `target_mac_address` | `[6]byte` / `uint8_t[6]` | `4,bytes` | MAC address of the frame's intended recipient |
| 18 | 6 | `LastHopMacAddress` / `last_hop_mac_address` | `[6]byte` / `uint8_t[6]` | `5,bytes` | MAC address of the node that last relayed the frame |
| 24 | 64 | `Data` / `data` | `[64]byte` / `uint8_t[64]` | `6,bytes,optional` | Message payload; for `MeshTypeSerialCmdBroadcast` frames, byte 0 is an opcode (see Opcodes) |
| 88 | 1 | `HopCount` / `hop_count` | `uint8` / `uint8_t` | `7,uint32` | Number of mesh hops the frame has traversed |
| 89 | 4 | `EpochNum` / `epoch_num` | `uint32` / `uint32_t` | `8,uint32` | Mesh epoch / topology generation number |
| 93 | 2 | `SeqNum` / `seq_num` | `uint16` / `uint16_t` | `9,uint32` | Per-origin sequence number |
| 95 | 32 | `EnrollmentPublicKey` / `enrollment_public_key` | `[32]byte` / `uint8_t[32]` | `11,bytes,optional,public_key` | Node's public key, sent during enrollment (`MeshTypeEnrollment`) |
| 127 | 1 | `RouteLen` / `route_len` | `uint8` / `uint8_t` | `12,uint32,optional` | v3: length (in hops) of `RoutePath` — downlink source route (Phase 3) / uplink path accumulator |
| 128 | 48 | `RoutePath` / `route_path` | `[48]byte` / `uint8_t[48]` | `13,bytes,optional` | v3: accumulated hop MAC path; 48 = MAX_HOPS(8) × 6 bytes/MAC |
| 176 | 16 | `AuthTag` / `auth_tag` | `[16]byte` / `uint8_t[16]` | `14,bytes,optional` | v3: ChaCha20-Poly1305 tag over `Data[64]` (E2E AEAD) |
| 192 | 8 | `AuthPath` / `auth_path` | `[8]byte` / `uint8_t[8]` | `17,bytes,optional,authPath` | v4: chained HMAC-SHA256-64 over the relay-accumulated `RoutePath` (Phase C, issue #44) |

**Total: 192 + 8 = 200 bytes.** This also matches the breakdown in the `WireSize` doc comment:
127 bytes (v2 prefix: `ProtoVersion` .. `EnrollmentPublicKey`) + 1 (`RouteLen`) + 48 (`RoutePath`)
+ 16 (`AuthTag`) + 8 (`AuthPath`) = 200. Field 127 is where `RouteLen` starts — the v2 prefix
fields (offsets 0–126) are unchanged by the v3/v4 additions, which is why v3 could append new
fields without breaking the layout of any field that existed before it.

Note the `proto` field numbers (10, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13, 14, 17) are not
sequential in wire order and skip 15 and 16 — this is transcribed as-is from `message/message.go`.
Fields 15 and 16 were `SecondaryMasterMac` and `SecondaryPublicKey`, removed in the `v0.6.0` wire
shrink; proto field numbers are never reused, so they remain permanently retired. See
[`docs/making_a_protocol_change.md`](making_a_protocol_change.md) for the full explanation.
