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

// WriteReadMetaReq writes a ReadMeta RPC packet to the remote end.
func (t *Transiever) WriteReadMetaReq(l *ReadMetaReq) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktReadMeta)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

// GetReadMetaReq decodes a ReadMetaReq packet from the network.
func (t *Transiever) GetReadMetaReq(l *ReadMetaReq) error {
	return t.packetDecoder.Decode(l)
}

// WriteReadMetaResp writes a ReadMetaResp RPC packet to the remote end.
func (t *Transiever) WriteReadMetaResp(l *ReadMetaResp) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktReadMetaResp)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

// GetReadMetaResp decodes a ReadMetaResp packet from the network.
func (t *Transiever) GetReadMetaResp(l *ReadMetaResp) error {
	return t.packetDecoder.Decode(l)
}

// WriteListReq writes a List RPC packet to the remote end.
func (t *Transiever) WriteListReq(l *ListReq) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktList)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

// GetListReq decodes a ListReq packet from the network.
func (t *Transiever) GetListReq(l *ListReq) error {
	return t.packetDecoder.Decode(l)
}

// WriteListResp writes a ListResp RPC packet to the remote end.
func (t *Transiever) WriteListResp(l *ListResp) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktListResp)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

// GetListResp decodes a ListResp packet from the network.
func (t *Transiever) GetListResp(l *ListResp) error {
	return t.packetDecoder.Decode(l)
}

// WriteFetchReq writes a Fetch RPC packet to the remote end.
func (t *Transiever) WriteFetchReq(l *FetchReq) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktFetch)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

// GetFetchReq decodes a FetchReq packet from the network.
func (t *Transiever) GetFetchReq(l *FetchReq) error {
	return t.packetDecoder.Decode(l)
}

// WriteFetchResp writes a FetchResp RPC packet to the remote end.
func (t *Transiever) WriteFetchResp(l *FetchResp) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktFetchResp)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

// GetFetchResp decodes a FetchResp packet from the network.
func (t *Transiever) GetFetchResp(l *FetchResp) error {
	return t.packetDecoder.Decode(l)
}

func (t *Transiever) WriteReadDataReq(l *ReadDataReq) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktReadData)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

func (t *Transiever) GetReadDataReq(l *ReadDataReq) error {
	return t.packetDecoder.Decode(l)
}

func (t *Transiever) WriteReadDataResp(l *ReadDataResp) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktReadDataResp)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

func (t *Transiever) GetReadDataResp(l *ReadDataResp) error {
	return t.packetDecoder.Decode(l)
}

func (t *Transiever) WriteStoreReq(l *StoreReq) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktStore)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

func (t *Transiever) GetStoreReq(l *StoreReq) error {
	return t.packetDecoder.Decode(l)
}

func (t *Transiever) WriteStoreResp(l *StoreResp) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktStoreResp)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

func (t *Transiever) GetStoreResp(l *StoreResp) error {
	return t.packetDecoder.Decode(l)
}

func (t *Transiever) WriteMkdirReq(l *MkdirReq) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktMkdir)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

func (t *Transiever) GetMkdirReq(l *MkdirReq) error {
	return t.packetDecoder.Decode(l)
}

func (t *Transiever) WriteMkdirResp(l *MkdirResp) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktMkdirResp)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

func (t *Transiever) GetMkdirResp(l *MkdirResp) error {
	return t.packetDecoder.Decode(l)
}

func (t *Transiever) WriteDeleteReq(l *DeleteReq) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktDelete)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

func (t *Transiever) GetDeleteReq(l *DeleteReq) error {
	return t.packetDecoder.Decode(l)
}

func (t *Transiever) WriteDeleteResp(l *DeleteResp) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()

	err := t.packetEncoder.Encode(PktDeleteResp)
	if err != nil {
		return err
	}
	return t.packetEncoder.Encode(l)
}

func (t *Transiever) GetDeleteResp(l *DeleteResp) error {
	return t.packetDecoder.Decode(l)
}
