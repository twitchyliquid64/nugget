package client

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/twitchyliquid64/nugget/logger"
	"github.com/twitchyliquid64/nugget/packet"
)

// RemoteSource represents a nuggFS endpoint over
// an authenticated network connection.
type RemoteSource struct {
	conn       *tls.Conn
	transiever *packet.Transiever
	logger     *logger.Logger

	shouldRun bool           //set to true if routines should run
	wg        sync.WaitGroup //tracks all routines

	onFatalChan chan error //if non-nil, fatal errors will be sent down it
	fatal       error      //set if there was a fatal error on this RemoteSource

	latency time.Duration //current latency - updated periodically by keepAliveRoutine
}

// Open starts a connection to the given nuggFS remote source using the
// certificate paths provided.
func Open(addr, certPemPath, keyPemPath, caCertPath string, l *logger.Logger, fatalErr chan error) (*RemoteSource, error) {
	conn, err := connect(addr, certPemPath, keyPemPath, caCertPath)
	if err != nil {
		return nil, err
	}

	rs := &RemoteSource{
		conn:        conn,
		shouldRun:   true,
		transiever:  packet.MakeTransiever(conn, conn),
		logger:      l,
		onFatalChan: fatalErr,
	}
	go rs.readServiceRoutine()
	go rs.keepAliveRoutine()
	return rs, nil
}

func connect(addr, certPemPath, keyPemPath, caCertPath string) (*tls.Conn, error) {
	tlsConf, err := tlsConfig(certPemPath, keyPemPath, caCertPath)
	if err != nil {
		return nil, err
	}
	conn, err := tls.Dial("tcp", addr, tlsConf)
	return conn, err
}

func (c *RemoteSource) keepAliveRoutine() {
	c.wg.Add(1)
	defer c.wg.Done()
	for c.shouldRun {
		c.ping()
		time.Sleep(time.Second * 2)
	}
}

func (c *RemoteSource) readServiceRoutine() {
	c.wg.Add(1)
	defer c.wg.Done()

	for c.shouldRun {
		pktType, err := c.transiever.Decode()
		if err != nil {
			c.logger.Error("net-read", err)
			c.fatalInternalError(err)
			return
		}

		var processingError error
		switch pktType {
		case packet.PktPong:
			var pong packet.PingPong
			processingError = c.transiever.GetPing(&pong)
			c.latency = time.Now().Sub(pong.Sent)
			fmt.Println("Latency: ", c.latency)
		}

		if processingError != nil {
			c.logger.Error("net-process", processingError)
			c.fatalInternalError(err)
			return
		}
	}
}

func (c *RemoteSource) fatalInternalError(err error) {
	c.shouldRun = false
	c.conn.Close()
	c.fatal = err
	if c.onFatalChan != nil {
		c.onFatalChan <- err
	}
}

// Ready returns true if the connection is healthy and ready for RPCs.
func (c *RemoteSource) Ready() bool {
	return c.shouldRun && c.fatal == nil
}

func (c *RemoteSource) ping() error {
	var ping packet.PingPong
	ping.Sent = time.Now()

	return c.transiever.WritePing(&ping)
}
