package serv

import (
	"net"

	"github.com/twitchyliquid64/nugget"
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
		case packet.PktReadData:
			processingError = c.processReadDataPkt(trans)
		case packet.PktStore:
			processingError = c.processStorePkt(trans)
		case packet.PktMkdir:
			processingError = c.processMkdirPkt(trans)
		case packet.PktDelete:
			processingError = c.processDeletePkt(trans)
		case packet.PktWrite:
			processingError = c.processWritePkt(trans)
		case packet.PktRead:
			processingError = c.processReadPkt(trans)
		}

		if processingError != nil {
			c.Manager.logger.Error("client-read", processingError)
			return
		}
	}
}

func (c *Duplex) processReadPkt(trans *packet.Transiever) error {
	var readRequest packet.ReadReq
	err := trans.GetReadReq(&readRequest)
	if err != nil {
		return err
	}
	c.Manager.logger.Info("client-read", "Got Read request for ", readRequest.Path)

	var readResponse packet.ReadResp
	readResponse.ID = readRequest.ID

	if c.Manager.isOptimisedProvider {
		p := c.Manager.provider.(nugget.OptimisedDataSourceSink)
		data, err := p.Read(readRequest.Path, readRequest.Offset, readRequest.Size)
		if err == nuggdb.ErrChunkNotFound || err == nuggdb.ErrMetaNotFound || err == nuggdb.ErrPathNotFound {
			readResponse.ErrorCode = packet.ErrNoEntity
		} else if err != nil {
			c.Manager.logger.Warning("client-read", "Read() error: ", err)
			readResponse.ErrorCode = packet.ErrUnspec
		}
		readResponse.Data = data

	} else {
		c.Manager.logger.Warning("client-read", "Provider is not optimized - falling back to Fetch/slice strategy.")
		_, _, data, err := c.Manager.provider.Fetch(readRequest.Path)
		if err != nil {
			if err == nuggdb.ErrChunkNotFound || err == nuggdb.ErrMetaNotFound || err == nuggdb.ErrPathNotFound {
				readResponse.ErrorCode = packet.ErrNoEntity
			} else {
				readResponse.ErrorCode = packet.ErrUnspec
			}
		} else {
			if readRequest.Offset > int64(len(data)) {
				data = nil
			} else {
				data = data[readRequest.Offset:]
			}
			if int64(len(data)) > readRequest.Size {
				data = data[:readRequest.Size]
			}
			readResponse.Data = data
		}
	}
	return trans.WriteReadResp(&readResponse)
}

func (c *Duplex) processWritePkt(trans *packet.Transiever) error {
	var writeRequest packet.WriteReq
	err := trans.GetWriteReq(&writeRequest)
	if err != nil {
		return err
	}
	c.Manager.logger.Info("client-read", "Got Write request for ", writeRequest.Path)

	var writeResponse packet.WriteResp
	writeResponse.ID = writeRequest.ID

	if c.Manager.isOptimisedProvider {
		p := c.Manager.provider.(nugget.OptimisedDataSourceSink)
		written, entryID, meta, err := p.Write(writeRequest.Path, writeRequest.Offset, writeRequest.Data)

		writeResponse.Written = written
		writeResponse.Meta = *(meta.(*nuggdb.EntryMetadata))
		writeResponse.EntryID = entryID
		if err != nil {
			if err == nuggdb.ErrChunkNotFound || err == nuggdb.ErrMetaNotFound || err == nuggdb.ErrPathNotFound {
				writeResponse.ErrorCode = packet.ErrNoEntity
			} else {
				writeResponse.ErrorCode = packet.ErrUnspec
			}
		}

	} else {
		c.Manager.logger.Warning("client-read", "Provider is not optimized - falling back to Fetch/Write.")
		_, _, data, err := c.Manager.provider.Fetch(writeRequest.Path)
		if err != nil {
			if err == nuggdb.ErrChunkNotFound || err == nuggdb.ErrMetaNotFound || err == nuggdb.ErrPathNotFound {
				writeResponse.ErrorCode = packet.ErrNoEntity
			} else {
				writeResponse.ErrorCode = packet.ErrUnspec
			}
		} else {
			newData := doWrite(writeRequest.Offset, writeRequest.Data, data)
			entryID, meta, err := c.Manager.provider.Store(writeRequest.Path, newData)
			writeResponse.Written = int64(len(writeRequest.Data))
			writeResponse.Meta = *(meta.(*nuggdb.EntryMetadata))
			writeResponse.EntryID = entryID
			if err != nil {
				if err == nuggdb.ErrChunkNotFound || err == nuggdb.ErrMetaNotFound || err == nuggdb.ErrPathNotFound {
					writeResponse.ErrorCode = packet.ErrNoEntity
				} else {
					writeResponse.ErrorCode = packet.ErrUnspec
				}
			}
		}
	}

	return trans.WriteWriteResp(&writeResponse)
}

// doWrite does the buffer manipulation to perform a write. Data buffers are kept
// contiguous.
// Credit: bwester (consulfs)
func doWrite(offset int64, writeData []byte, fileData []byte) []byte {
	fileEnd := int64(len(fileData))
	writeEnd := offset + int64(len(writeData))
	var buf []byte
	if writeEnd > fileEnd {
		buf = make([]byte, writeEnd)
		if fileEnd <= offset {
			copy(buf, fileData)
		} else {
			copy(buf, fileData[:offset])
		}
	} else {
		buf = fileData
	}
	copy(buf[offset:writeEnd], writeData)
	return buf
}

func (c *Duplex) processDeletePkt(trans *packet.Transiever) error {
	var deleteRequest packet.DeleteReq
	err := trans.GetDeleteReq(&deleteRequest)
	if err != nil {
		return err
	}
	c.Manager.logger.Info("client-read", "Got Delete request for ", deleteRequest.Path)

	var deleteResponse packet.DeleteResp
	deleteResponse.ID = deleteRequest.ID
	err = c.Manager.provider.Delete(deleteRequest.Path)
	if err != nil {
		if err == nuggdb.ErrChunkNotFound || err == nuggdb.ErrMetaNotFound || err == nuggdb.ErrPathNotFound {
			deleteResponse.ErrorCode = packet.ErrNoEntity
		} else {
			deleteResponse.ErrorCode = packet.ErrUnspec
		}
	}

	return trans.WriteDeleteResp(&deleteResponse)
}

func (c *Duplex) processMkdirPkt(trans *packet.Transiever) error {
	var mkdirRequest packet.MkdirReq
	err := trans.GetMkdirReq(&mkdirRequest)
	if err != nil {
		return err
	}
	c.Manager.logger.Info("client-read", "Got Mkdir request for ", mkdirRequest.Path)

	var mkdirResponse packet.MkdirResp
	mkdirResponse.ID = mkdirRequest.ID
	entryID, meta, err := c.Manager.provider.Mkdir(mkdirRequest.Path)
	mkdirResponse.EntryID = entryID
	mkdirResponse.Meta = *(meta.(*nuggdb.EntryMetadata))
	if err != nil {
		if err == nuggdb.ErrChunkNotFound || err == nuggdb.ErrMetaNotFound || err == nuggdb.ErrPathNotFound {
			mkdirResponse.ErrorCode = packet.ErrNoEntity
		} else {
			mkdirResponse.ErrorCode = packet.ErrUnspec
		}
	}

	return trans.WriteMkdirResp(&mkdirResponse)
}

func (c *Duplex) processStorePkt(trans *packet.Transiever) error {
	var storeRequest packet.StoreReq
	err := trans.GetStoreReq(&storeRequest)
	if err != nil {
		return err
	}
	c.Manager.logger.Info("client-read", "Got Store request for ", storeRequest.Path)

	var storeResponse packet.StoreResp
	storeResponse.ID = storeRequest.ID
	entryID, meta, err := c.Manager.provider.Store(storeRequest.Path, storeRequest.Data)
	storeResponse.EntryID = entryID
	storeResponse.Meta = *(meta.(*nuggdb.EntryMetadata))
	if err != nil {
		storeResponse.ErrorCode = packet.ErrUnspec
	}

	return trans.WriteStoreResp(&storeResponse)
}

func (c *Duplex) processReadDataPkt(trans *packet.Transiever) error {
	var readDataRequest packet.ReadDataReq
	err := trans.GetReadDataReq(&readDataRequest)
	if err != nil {
		return err
	}
	c.Manager.logger.Info("client-read", "Got ReadData request for ", readDataRequest.ChunkID)

	var readDataResponse packet.ReadDataResp
	readDataResponse.ID = readDataRequest.ID
	d, err := c.Manager.provider.ReadData(readDataRequest.ChunkID)
	readDataResponse.Data = d
	if err != nil {
		if err == nuggdb.ErrChunkNotFound {
			readDataResponse.ErrorCode = packet.ErrNoEntity
		} else {
			readDataResponse.ErrorCode = packet.ErrUnspec
		}
	}

	return trans.WriteReadDataResp(&readDataResponse)
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
