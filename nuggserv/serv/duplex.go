package serv

import (
	"fmt"
	"net"
	"os"

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
		t, err := trans.Decode()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Client read err: %s\n", err)
			return
		}
		fmt.Println(t)
	}
}

// ClientWriteLoop is the routine responsible for encoding packets and transmitting
// them to the remote end.
func (c *Duplex) ClientWriteLoop() {

}
