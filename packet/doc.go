package packet

import (
	"encoding/gob"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/twitchyliquid64/nugget"
)

// packet contains structs for representing data on the wire, along with serializing / deserializing it.

// PktType represents the type of a packet on the wire.
type PktType byte

// Packet types
const (
	PktUnknown PktType = iota
	PktPing
	PktPong
	PktLookup
	PktLookupResp
)

// ErrorCode represents classes of RPC failures.
type ErrorCode byte

// Error codes
const (
	ErrNoError ErrorCode = iota
	ErrNoEntity
	ErrIOErr
	ErrTimeout
	ErrUnspec
)

// PingPong represents a ping/pong packet on the wire
type PingPong struct {
	Sent time.Time
}

// LookupReq represent a lookup RPC on the wire
type LookupReq struct {
	ID   uint64
	Path string
}

// LookupResp represents the response to a lookup RPC on the wire
type LookupResp struct {
	ID        uint64
	EntryID   nugget.EntryID
	ErrorCode ErrorCode
	ErrorText string
}

// Transiever takes a network bytestream and interprets it into packet structures.
type Transiever struct {
	packetDecoder *gob.Decoder
	packetEncoder *gob.Encoder

	reader io.Reader
	writer io.Writer

	sendLock sync.Mutex
}

// ErrorCodeToErr maps error codes returned via RPC to actual error types.
func ErrorCodeToErr(code ErrorCode) error {
	switch code {
	case ErrNoError:
		return nil
	case ErrNoEntity:
		return errors.New("No entity")
	case ErrIOErr:
		return errors.New("IO Error")
	case ErrTimeout:
		return errors.New("Timeout")
	case ErrUnspec:
		return errors.New("Unspecified")
	}
	return errors.New("Unknown Error")
}
