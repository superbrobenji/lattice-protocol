// Package message defines the Lattice mesh message type constants.
// C headers in c/ are generated from these constants — run "go generate ./..." to regenerate.
//
//go:generate go run ../cmd/gen-headers/main.go
package message

const (
	MeshTypeAdapterData        = uint8(0) // node data relayed toward server
	MeshTypeMasterBeacon       = uint8(1) // master→mesh: topology beacon
	MeshTypeEnrollment         = uint8(2) // node→master: enrollment request
	MeshTypeSerialCmdBroadcast = uint8(3) // server→node: serial command broadcast
	MeshTypeJoinAck            = uint8(4) // server→node: enrollment approved
	MeshTypeRouteReport        = uint8(5) // node→server: routing path report
)
