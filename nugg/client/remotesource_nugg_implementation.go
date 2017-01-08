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
