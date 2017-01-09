package packet

import (
	"encoding/gob"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/twitchyliquid64/nugget"
	"github.com/twitchyliquid64/nugget/nuggdb"
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
	PktReadMeta
	PktReadMetaResp
	PktList
	PktListResp
	PktFetch
	PktFetchResp
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
}

// ReadMetaReq represents a ReadMeta RPC on the wire
type ReadMetaReq struct {
	ID      uint64
	EntryID nugget.EntryID
}

// ReadMetaResp respresents the response to a ReadMeta RPC on the wire
type ReadMetaResp struct {
	ID        uint64
	ErrorCode ErrorCode
	Meta      nuggdb.EntryMetadata
}

// ListReq represents a List RPC on the wire
type ListReq struct {
	ID   uint64
	Path string
}

// ListResp represents the response to a List RPC on the wire
type ListResp struct {
	ID        uint64
	ErrorCode ErrorCode
	Entries   []nuggdb.DirEntry
}

// FetchReq represents a Fetch RPC on the wire
type FetchReq struct {
	ID        uint64
	ErrorCode ErrorCode
	Path      string
}

// FetchResp represents the response to a Fetch RPC on the wire
type FetchResp struct {
	ID        uint64
	ErrorCode ErrorCode
	EntryID   nugget.EntryID
	Meta      nuggdb.EntryMetadata
	Data      []byte
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
