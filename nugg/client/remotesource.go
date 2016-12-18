package client

import "crypto/tls"

// RemoteSource represents a nuggFS endpoint over
// an authenticated network connection.
type RemoteSource struct {
	conn *tls.Conn
}

func connect(addr, certPemPath, keyPemPath, caCertPath string) (*tls.Conn, error) {
	tlsConf, err := tlsConfig(certPemPath, keyPemPath, caCertPath)
	if err != nil {
		return nil, err
	}
	conn, err := tls.Dial("tcp", addr, tlsConf)
	return conn, err
}

// Open starts a connection to the given nuggFS remote source using the
// certificate paths provided.
func Open(addr, certPemPath, keyPemPath, caCertPath string) (*RemoteSource, error) {
	conn, err := connect(addr, certPemPath, keyPemPath, caCertPath)
	if err != nil {
		return nil, err
	}

	return &RemoteSource{
		conn: conn,
	}, nil
}
