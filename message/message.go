// Package message defines the Lattice mesh wire-format message struct.
// Struct tags drive codegen: `c` → packed C struct field type, `proto` → "fieldNum,protoType[,optional][,protoName]".
// Run "go generate ./..." to regenerate c/mesh_message.h and proto/mesh.proto.
package message

// MeshMessage is the 250-byte packed wire-format frame for the Lattice mesh (protocol v4).
// Field order matches the packed C struct — do not reorder without updating the static_assert.
// v3 additions (route_len..secondary_public_key) append after enrollment_public_key so the
// 127-byte v2 prefix layout is unchanged.
type MeshMessage struct {
	ProtoVersion        uint8    `c:"uint8_t"     proto:"10,uint32"`
	MessageType         uint8    `c:"uint8_t"     proto:"1,uint32"`
	DataType            int32    `c:"int32_t"     proto:"2,sint32"`
	OriginMacAddress    [6]byte  `c:"uint8_t[6]"  proto:"3,bytes"`
	TargetMacAddress    [6]byte  `c:"uint8_t[6]"  proto:"4,bytes"`
	LastHopMacAddress   [6]byte  `c:"uint8_t[6]"  proto:"5,bytes"`
	Data                [64]byte `c:"uint8_t[64]" proto:"6,bytes,optional"`
	HopCount            uint8    `c:"uint8_t"     proto:"7,uint32"`
	EpochNum            uint32   `c:"uint32_t"    proto:"8,uint32"`
	SeqNum              uint16   `c:"uint16_t"    proto:"9,uint32"`
	EnrollmentPublicKey [32]byte `c:"uint8_t[32]" proto:"11,bytes,optional,public_key"`
	// v3: downlink source route (Phase 3) / uplink path accumulator. 60 = MAX_HOPS(10) × 6.
	RouteLen  uint8    `c:"uint8_t"     proto:"12,uint32,optional"`
	RoutePath [60]byte `c:"uint8_t[60]" proto:"13,bytes,optional"`
	// v3: ChaCha20-Poly1305 tag over data[64] (E2E AEAD).
	AuthTag [16]byte `c:"uint8_t[16]" proto:"14,bytes,optional"`
	// v3: dual-master provisioning in JOIN_ACK (Phase 4). Zero elsewhere.
	SecondaryMasterMac  [6]byte  `c:"uint8_t[6]"  proto:"15,bytes,optional"`
	SecondaryPublicKey  [32]byte `c:"uint8_t[32]" proto:"16,bytes,optional"`
	// v4: chained HMAC-SHA256-64 over the relay-accumulated route_path (Phase C, issue #44).
	AuthPath [8]byte `c:"uint8_t[8]" proto:"17,bytes,optional,authPath"`
}

// WireSize is the expected packed byte size — enforced by static_assert in the generated C header.
// 127 (v2 prefix) + 1 + 60 + 16 + 6 + 32 + 8 = 250. Must stay ≤ 250 (ESP-NOW frame limit).
const WireSize = 250
