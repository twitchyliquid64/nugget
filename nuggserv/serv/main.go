package serv

import (
	"crypto/tls"
	"log"
	"net"

	"github.com/twitchyliquid64/nugget/nuggserv/remoteconn"
)

func (m *Manager) mainloop(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Listener err: ", err)
			break
		} else {
			remoteConn := initClient(conn, m)
			go remoteConn.ClientReadLoop()
			go remoteConn.ClientWriteLoop()
		}
	}
}

func initNetwork(listenAddr, certPemPath, keyPemPath, caCertPath string) (net.Listener, error) {
	tlsConf, err := tlsConfig(certPemPath, keyPemPath, caCertPath)
	if err != nil {
		return nil, err
	}
	listener, err := tls.Listen("tcp", listenAddr, tlsConf)
	return listener, err
}

func initClient(conn net.Conn, manager *Manager) *remoteconn.Duplex {
	return &remoteconn.Duplex{
		Conn:    conn,
		Manager: manager,
	}
}
