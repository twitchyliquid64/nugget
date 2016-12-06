package remoteconn

import (
	"net"

	"github.com/twitchyliquid64/nugget/nuggserv/serv"
)

type Duplex struct {
	Conn    net.Conn
	Manager *serv.Manager
}

func (c *Duplex) ClientReadLoop() {

}

func (c *Duplex) ClientWriteLoop() {

}
