package serv

import "net"

// Manager is the concrete type representing the network side of a server,
// and managing client connections.
type Manager struct {
	isOnline bool
	listener net.Listener
}
