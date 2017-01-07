package serv

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/twitchyliquid64/nugget/logger"
)

func (m *Manager) mainloop() {
	var tcpListener *net.TCPListener
	tcpListener, _ = m.listener.(*net.TCPListener)
	m.wg.Add(1)
	defer m.wg.Done()
	for m.isOnline {
		if tcpListener != nil {
			tcpListener.SetDeadline(time.Now().Add(time.Millisecond * 300))
		}
		conn, err := m.listener.Accept()
		if err != nil {
			m.logger.Error("listen", err)
			break
		} else {
			remoteConn := initClient(conn, m)
			go remoteConn.ClientReadLoop()
			go remoteConn.ClientWriteLoop()
		}
	}
}

// Close shuts down the listener and accept routine.
func (m *Manager) Close() error {
	m.isOnline = false
	m.wg.Done()
	return m.listener.Close()
}

func initNetwork(listenAddr, certPemPath, keyPemPath, caCertPath string) (net.Listener, error) {
	tlsConf, err := tlsConfig(certPemPath, keyPemPath, caCertPath)
	if err != nil {
		return nil, err
	}
	listener, err := tls.Listen("tcp", listenAddr, tlsConf)
	return listener, err
}

// NewServer initializes a network server on listeAddr, accepting connections which can authenticate themselves
// as based of the certificate at caCertPath. The TLS server authenticates itself using the cert/key at
// certPemPath and keyPemPath respectively.
func NewServer(listenAddr, certPemPath, keyPemPath, caCertPath string, logger *logger.Logger) (*Manager, error) {
	listener, err := initNetwork(listenAddr, certPemPath, keyPemPath, caCertPath)
	if err != nil {
		return nil, err
	}

	m := &Manager{
		isOnline: true,
		listener: listener,
		logger:   logger,
	}

	go m.mainloop()
	return m, nil
}

func initClient(conn net.Conn, manager *Manager) *Duplex {
	return &Duplex{
		Conn:    conn,
		Manager: manager,
	}
}
