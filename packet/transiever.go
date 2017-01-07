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
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktPing)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(ping)
}

// WritePong writes a ping packet to the remote end.
func (t *Transiever) WritePong(ping *PingPong) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktPong)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(ping)
}

// WriteLookupReq writes a Lookup RPC packet to the remote end.
func (t *Transiever) WriteLookupReq(l *LookupReq) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktLookup)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

// GetLookupReq decodes a LookupReq packet from the network.
func (t *Transiever) GetLookupReq(l *LookupReq) error {
	return t.packetDecoder.Decode(l)
}

// WriteLookupResp writes the response to a Lookup RPC packet to the remote end.
func (t *Transiever) WriteLookupResp(l *LookupResp) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktLookupResp)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

// GetLookupResp decodes a LookupReq packet from the network.
func (t *Transiever) GetLookupResp(l *LookupResp) error {
	return t.packetDecoder.Decode(l)
}
