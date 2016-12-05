package packet

import (
	"bytes"
	"encoding/gob"
	"io"
	"time"
)

// packet contains structs for representing data on the wire, along with serializing / deserializing it.

// PktType represents the type of a packet on the wire.
type PktType byte

// Packet types
const (
	PktUnknown PktType = iota
	PktPing
	PktPong
)

// Packet represents a data packet on the wire.
type Packet struct {
	Type PktType
	Data []byte
}

// PingPing represents a ping/pong packet on the wire
type PingPing struct {
	Sent time.Time
}

// Transiever takes a network bytestream and interprets it into packet structures.
type Transiever struct {
	justReadPacket Packet

	packetDecoder      *gob.Decoder
	packetEncoder      *gob.Encoder
	pingPongPktDecoder *gob.Decoder
	pingPongPktEncoder *gob.Encoder

	encapPktBufferInput  bytes.Buffer
	encapPktBufferOutput bytes.Buffer

	reader io.Reader
	writer io.Writer
}
