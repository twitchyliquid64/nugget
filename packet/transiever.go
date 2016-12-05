package packet

import (
	"encoding/gob"
	"io"
)

// TODO: GET RID OF THIS DANCE - JUST SEND THE TYPE THEN VALUE DOWN ONE ENCODER. Reverse on the other side in the same fashion, just without the extra encoders / decoders

// MakeTransiever returns an initialized Transiever object to be used for packet generation / decoding.
func MakeTransiever(reader io.Reader, writer io.Writer) *Transiever {
	ret := &Transiever{
		reader: reader,
		writer: writer,
	}
	ret.pingPongPktDecoder = gob.NewDecoder(&ret.encapPktBufferInput)
	ret.pingPongPktEncoder = gob.NewEncoder(&ret.encapPktBufferOutput)
	ret.packetDecoder = gob.NewDecoder(reader)
	ret.packetEncoder = gob.NewEncoder(writer)
	return ret
}

// Decode reads a single packet from the reader, returning its type and populating the packet temporary buffer. Future
// Invocations can get the packet-specific data using GetPing() etc.
func (t *Transiever) Decode() (PktType, error) {
	err := t.packetDecoder.Decode(&t.justReadPacket)
	if err != nil {
		return PktUnknown, err
	}

	t.encapPktBufferInput.Reset()
	t.encapPktBufferInput.Write(t.justReadPacket.Data)
	return t.justReadPacket.Type, nil
}

// GetPing decodes a ping packet from the temporary buffer.
func (t *Transiever) GetPing(ping *PingPing) error {
	return t.pingPongPktDecoder.Decode(ping)
}

// WritePing writes a ping packet to the remote end.
func (t *Transiever) WritePing(ping *PingPing) error {
	t.encapPktBufferOutput.Reset()
	e := t.pingPongPktEncoder.Encode(ping)
	if e != nil {
		return e
	}

	return t.packetEncoder.Encode(&Packet{
		Type: PktPing,
		Data: t.encapPktBufferOutput.Bytes(),
	})
}
