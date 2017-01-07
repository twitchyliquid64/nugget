package packet

import (
	"encoding/gob"
	"io"
	"sync"
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

// PingPong represents a ping/pong packet on the wire
type PingPong struct {
	Sent time.Time
}

// Transiever takes a network bytestream and interprets it into packet structures.
type Transiever struct {
	packetDecoder *gob.Decoder
	packetEncoder *gob.Encoder

	reader io.Reader
	writer io.Writer

	sendLock sync.Mutex
}
