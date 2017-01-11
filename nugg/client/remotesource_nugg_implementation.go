package client

import (
	"errors"
	"time"

	"github.com/twitchyliquid64/nugget"
	"github.com/twitchyliquid64/nugget/packet"
)

//ErrTimeout is returned if the remote server did not respond in time
var ErrTimeout = errors.New("Timeout waiting for response")

//ErrNotImplemented is returned if things are not yet implemented
var ErrNotImplemented = errors.New("Not implemented")

const defaultTimeout = time.Second * 4

// Lookup implements nugget.DataSource
func (c *RemoteSource) Lookup(path string) (nugget.EntryID, error) {
	responseChan := make(chan interface{})
	call := c.registerRPC(responseChan)
	defer c.unregisterRPC(call)

	var lookupRequest packet.LookupReq
	lookupRequest.ID = call.id
	lookupRequest.Path = path
	c.transiever.WriteLookupReq(&lookupRequest)

	select {
	case <-time.After(defaultTimeout):
		return nugget.EntryID{}, ErrTimeout
	case r := <-responseChan:
		lookupResp := r.(packet.LookupResp)
		if lookupResp.ErrorCode != packet.ErrNoError {
			return nugget.EntryID{}, packet.ErrorCodeToErr(lookupResp.ErrorCode)
		}
		return lookupResp.EntryID, nil
	}
}

// ReadMeta implements nugget.DataSource
func (c *RemoteSource) ReadMeta(entry nugget.EntryID) (nugget.NodeMetadata, error) {
	responseChan := make(chan interface{})
	call := c.registerRPC(responseChan)
	defer c.unregisterRPC(call)

	var readMetaRequest packet.ReadMetaReq
	readMetaRequest.ID = call.id
	readMetaRequest.EntryID = entry
	c.transiever.WriteReadMetaReq(&readMetaRequest)

	select {
	case <-time.After(defaultTimeout):
		return nil, ErrTimeout
	case r := <-responseChan:
		readMetaResp := r.(packet.ReadMetaResp)
		if readMetaResp.ErrorCode != packet.ErrNoError {
			return nil, packet.ErrorCodeToErr(readMetaResp.ErrorCode)
		}
		return &readMetaResp.Meta, nil
	}
}

// List implements nugget.DataSource
func (c *RemoteSource) List(path string) ([]nugget.DirEntry, error) {
	responseChan := make(chan interface{})
	call := c.registerRPC(responseChan)
	defer c.unregisterRPC(call)

	var listRequest packet.ListReq
	listRequest.ID = call.id
	listRequest.Path = path
	c.transiever.WriteListReq(&listRequest)

	select {
	case <-time.After(defaultTimeout):
		return nil, ErrTimeout
	case r := <-responseChan:
		listResp := r.(packet.ListResp)
		if listResp.ErrorCode != packet.ErrNoError {
			return nil, packet.ErrorCodeToErr(listResp.ErrorCode)
		}

		b := make([]nugget.DirEntry, len(listResp.Entries))
		for i := range listResp.Entries {
			b[i] = &listResp.Entries[i]
		}
		return b, nil
	}
}

// ReadData implements nugget.DataSource
func (c *RemoteSource) ReadData(node nugget.ChunkID) ([]byte, error) {
	return []byte(""), ErrNotImplemented
}

// Fetch implements nugget.DataSource
func (c *RemoteSource) Fetch(path string) (nugget.EntryID, nugget.NodeMetadata, []byte, error) {
	responseChan := make(chan interface{})
	call := c.registerRPC(responseChan)
	defer c.unregisterRPC(call)

	var fetchRequest packet.FetchReq
	fetchRequest.ID = call.id
	fetchRequest.Path = path
	c.transiever.WriteFetchReq(&fetchRequest)

	select {
	case <-time.After(defaultTimeout):
		return nugget.EntryID{}, nil, []byte(""), ErrTimeout
	case r := <-responseChan:
		fetchResp := r.(packet.FetchResp)
		if fetchResp.ErrorCode != packet.ErrNoError {
			return fetchResp.EntryID, &fetchResp.Meta, fetchResp.Data, packet.ErrorCodeToErr(fetchResp.ErrorCode)
		}
		return fetchResp.EntryID, &fetchResp.Meta, fetchResp.Data, nil
	}
}

// Store implements nugget.DataSink
func (c *RemoteSource) Store(path string, data []byte) (nugget.EntryID, nugget.NodeMetadata, error) {
	responseChan := make(chan interface{})
	call := c.registerRPC(responseChan)
	defer c.unregisterRPC(call)

	var storeRequest packet.StoreReq
	storeRequest.ID = call.id
	storeRequest.Path = path
	storeRequest.Data = data
	c.transiever.WriteStoreReq(&storeRequest)

	select {
	case <-time.After(defaultTimeout):
		return nugget.EntryID{}, nil, ErrTimeout
	case r := <-responseChan:
		storeResp := r.(packet.StoreResp)
		if storeResp.ErrorCode != packet.ErrNoError {
			return nugget.EntryID{}, nil, packet.ErrorCodeToErr(storeResp.ErrorCode)
		}
		return storeResp.EntryID, &storeResp.Meta, nil
	}
}

// Mkdir implements nugget.DataSink
func (c *RemoteSource) Mkdir(path string) (nugget.EntryID, nugget.NodeMetadata, error) {
	responseChan := make(chan interface{})
	call := c.registerRPC(responseChan)
	defer c.unregisterRPC(call)

	var mkdirRequest packet.MkdirReq
	mkdirRequest.ID = call.id
	mkdirRequest.Path = path
	c.transiever.WriteMkdirReq(&mkdirRequest)

	select {
	case <-time.After(defaultTimeout):
		return nugget.EntryID{}, nil, ErrTimeout
	case r := <-responseChan:
		mkdirResp := r.(packet.MkdirResp)
		if mkdirResp.ErrorCode != packet.ErrNoError {
			return nugget.EntryID{}, nil, packet.ErrorCodeToErr(mkdirResp.ErrorCode)
		}
		return mkdirResp.EntryID, &mkdirResp.Meta, nil
	}
}

// Delete implements nugget.DataSink
func (c *RemoteSource) Delete(path string) error {
	responseChan := make(chan interface{})
	call := c.registerRPC(responseChan)
	defer c.unregisterRPC(call)

	var deleteRequest packet.DeleteReq
	deleteRequest.ID = call.id
	deleteRequest.Path = path
	c.transiever.WriteDeleteReq(&deleteRequest)

	select {
	case <-time.After(defaultTimeout):
		return ErrTimeout
	case r := <-responseChan:
		deleteResp := r.(packet.DeleteResp)
		if deleteResp.ErrorCode != packet.ErrNoError {
			return packet.ErrorCodeToErr(deleteResp.ErrorCode)
		}
		return nil
	}
}

// Close implements nugget.DataSink
func (c *RemoteSource) Close() error {
	c.conn.Close()
	return ErrNotImplemented
}
