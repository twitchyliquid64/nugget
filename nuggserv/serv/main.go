package serv

import (
	"crypto/tls"
	"log"
	"net"
)

func mainloop(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		log.Println("Listener err: ", err)
		break
	} else {
		// Do shit with the connection?
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

func initClient() {

}

func clientReadLoop() {

}

func clientWriteLoop() {

}
