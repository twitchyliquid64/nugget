package serv

import (
	"net"

	"github.com/twitchyliquid64/nugget/nuggdb"
	"github.com/twitchyliquid64/nugget/packet"
)

// Duplex is the concrete type representing a authenticated link
// to a client.
type Duplex struct {
	Conn    net.Conn
	Manager *Manager
}

// ClientReadLoop is the routine responsible for recieving and decoding packets
// from the remote end.
func (c *Duplex) ClientReadLoop() {
	trans := packet.MakeTransiever(c.Conn, c.Conn)
	for {
		pktType, err := trans.Decode()
		if err != nil {
			c.Manager.logger.Warning("client-read", err)
			return
		}

		var processingError error
		switch pktType {
		case packet.PktPing:
			processingError = c.processPingPkt(trans)
		case packet.PktLookup:
			processingError = c.processLookupPkt(trans)
		case packet.PktReadMeta:
			processingError = c.processReadMetaPkt(trans)
		case packet.PktList:
			processingError = c.processListPkt(trans)
		case packet.PktFetch:
			processingError = c.processFetchPkt(trans)
		}

		if processingError != nil {
			c.Manager.logger.Error("client-read", processingError)
			return
		}
	}
}

func (c *Duplex) processListPkt(trans *packet.Transiever) error {
	var listRequest packet.ListReq
	err := trans.GetListReq(&listRequest)
	if err != nil {
		return err
	}
	c.Manager.logger.Info("client-read", "Got List request for ", listRequest.Path)

	var listResponse packet.ListResp
	listResponse.ID = listRequest.ID
	entries, err := c.Manager.provider.List(listRequest.Path)
	if err != nil {
		if err == nuggdb.ErrPathNotFound {
			listResponse.ErrorCode = packet.ErrNoEntity
		} else {
			listResponse.ErrorCode = packet.ErrUnspec
		}
	} else {
		b := make([]nuggdb.DirEntry, len(entries))
		for i := range entries {
			b[i] = *(entries[i].(*nuggdb.DirEntry))
		}
		listResponse.Entries = b
	}

	return trans.WriteListResp(&listResponse)
}

func (c *Duplex) processReadMetaPkt(trans *packet.Transiever) error {
	var readMetaRequest packet.ReadMetaReq
	err := trans.GetReadMetaReq(&readMetaRequest)
	if err != nil {
		return err
	}
	c.Manager.logger.Info("client-read", "Got ReadMeta request for ", readMetaRequest.EntryID)

	var readMetaResponse packet.ReadMetaResp
	readMetaResponse.ID = readMetaRequest.ID

	meta, err := c.Manager.provider.ReadMeta(readMetaRequest.EntryID)
	if err != nil {
		if err == nuggdb.ErrMetaNotFound {
			readMetaResponse.ErrorCode = packet.ErrNoEntity
		} else {
			readMetaResponse.ErrorCode = packet.ErrUnspec
		}
	} else {
		readMetaResponse.Meta = *(meta.(*nuggdb.EntryMetadata))
	}

	return trans.WriteReadMetaResp(&readMetaResponse)
}

func (c *Duplex) processFetchPkt(trans *packet.Transiever) error {
	var fetchReq packet.FetchReq
	err := trans.GetFetchReq(&fetchReq)
	if err != nil {
		return err
	}
	c.Manager.logger.Info("client-read", "Got Fetch request for ", fetchReq.Path)

	var fetchResponse packet.FetchResp
	fetchResponse.ID = fetchReq.ID

	entryID, metadata, data, err := c.Manager.provider.Fetch(fetchReq.Path)
	fetchResponse.Meta = *(metadata.(*nuggdb.EntryMetadata))
	fetchResponse.Data = data
	fetchResponse.EntryID = entryID
	if err != nil {
		if err == nuggdb.ErrMetaNotFound || err == nuggdb.ErrPathNotFound {
			fetchResponse.ErrorCode = packet.ErrNoEntity
		} else {
			fetchResponse.ErrorCode = packet.ErrUnspec
		}
	}

	return trans.WriteFetchResp(&fetchResponse)
}

func (c *Duplex) processLookupPkt(trans *packet.Transiever) error {
	var lookupRequest packet.LookupReq
	err := trans.GetLookupReq(&lookupRequest)
	if err != nil {
		return err
	}
	c.Manager.logger.Info("client-read", "Got Lookup request for ", lookupRequest.Path)

	var lookupResponse packet.LookupResp
	lookupResponse.ID = lookupRequest.ID

	lookupResponse.EntryID, err = c.Manager.provider.Lookup(lookupRequest.Path)
	if err != nil {
		if err == nuggdb.ErrPathNotFound {
			lookupResponse.ErrorCode = packet.ErrNoEntity
		} else {
			lookupResponse.ErrorCode = packet.ErrUnspec
		}
	}

	return trans.WriteLookupResp(&lookupResponse)
}

func (c *Duplex) processPingPkt(trans *packet.Transiever) error {
	var err error
	var ping packet.PingPong

	err = trans.GetPing(&ping)
	if err != nil {
		return err
	}
	return trans.WritePong(&ping)
}
