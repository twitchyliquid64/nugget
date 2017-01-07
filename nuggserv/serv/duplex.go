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
		}

		if processingError != nil {
			c.Manager.logger.Error("client-read", processingError)
			return
		}
	}
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

	err = trans.WriteLookupResp(&lookupResponse)
	if err != nil {
		return err
	}
	return nil
}

func (c *Duplex) processPingPkt(trans *packet.Transiever) error {
	var err error
	var ping packet.PingPong

	err = trans.GetPing(&ping)
	if err != nil {
		return err
	}
	err = trans.WritePong(&ping)
	if err != nil {
		return err
	}
	return nil
}
