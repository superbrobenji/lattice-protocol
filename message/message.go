// Package message defines the Lattice mesh wire-format message struct.
// Struct tags drive codegen: `c` → packed C struct field type, `proto` → "fieldNum,protoType[,optional][,protoName]".
// Run "go generate ./..." to regenerate c/mesh_message.h and proto/mesh.proto.
package message

// MeshMessage is the 127-byte packed wire-format frame for the Lattice mesh.
// Field order matches the packed C struct — do not reorder without updating the static_assert.
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
}

// WireSize is the expected packed byte size — enforced by static_assert in the generated C header.
const WireSize = 127
