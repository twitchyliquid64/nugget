package serv

import (
	"net"

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
		}

		if processingError != nil {
			c.Manager.logger.Error("client-read", processingError)
			return
		}
	}
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

// ClientWriteLoop is the routine responsible for encoding packets and transmitting
// them to the remote end. TODO: Change it to the routine servicing a response write channel?
func (c *Duplex) ClientWriteLoop() {
}
