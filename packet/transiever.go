package packet

import (
	"encoding/gob"
	"io"
)

// MakeTransiever returns an initialized Transiever object to be used for packet generation / decoding.
func MakeTransiever(reader io.Reader, writer io.Writer) *Transiever {
	ret := &Transiever{
		reader: reader,
		writer: writer,
	}
	ret.packetDecoder = gob.NewDecoder(reader)
	ret.packetEncoder = gob.NewEncoder(writer)
	return ret
}

// Decode reads the type prefix of the next packet, leaving the actual packet in the buffer.
// Future invocations (based on the type) can get the packet-specific data using GetPing() etc.
func (t *Transiever) Decode() (PktType, error) {
	var pktType PktType
	err := t.packetDecoder.Decode(&pktType)
	if err != nil {
		return PktUnknown, err
	}
	return pktType, nil
}

// GetPing decodes a ping/pong packet from the network.
func (t *Transiever) GetPing(ping *PingPong) error {
	return t.packetDecoder.Decode(ping)
}

// WritePing writes a ping packet to the remote end.
func (t *Transiever) WritePing(ping *PingPong) error {
	err := t.packetEncoder.Encode(PktPing)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(ping)
}
