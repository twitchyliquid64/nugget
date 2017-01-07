package client

import (
	"crypto/rand"
	"encoding/binary"
)

// Call represents an in-flight RPC.
type Call struct {
	id           uint64
	responseChan chan interface{}
}

func (c *RemoteSource) dispatchCallResponse(id uint64, data interface{}) {
	c.pendingLock.Lock()
	defer c.pendingLock.Unlock()

	call, ok := c.pending[id]
	if !ok {
		c.logger.Warning("rpc-response", "Could not match RPC response ", id, " with tracked request")
	} else {
		call.responseChan <- data
	}
}

func (c *RemoteSource) registerRPC(ch chan interface{}) *Call {
	c.pendingLock.Lock()
	defer c.pendingLock.Unlock()
	id := getRandInt()
	call := &Call{id: id, responseChan: ch}
	c.pending[id] = call
	return call
}

func (c *RemoteSource) unregisterRPC(call *Call) {
	c.pendingLock.Lock()
	defer c.pendingLock.Unlock()
	delete(c.pending, call.id)
}

func getRandInt() uint64 {
	b := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	if _, err := rand.Reader.Read(b); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint64(b)
}
