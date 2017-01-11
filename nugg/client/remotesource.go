package client

import (
	"crypto/tls"
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

	pendingLock sync.Mutex
	pending     map[uint64]*Call
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
		pending:     map[uint64]*Call{},
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

		case packet.PktLookupResp:
			processingError = c.processLookupResponse()

		case packet.PktReadMetaResp:
			processingError = c.processReadMetaResponse()

		case packet.PktListResp:
			processingError = c.processListResponse()

		case packet.PktFetchResp:
			processingError = c.processFetchResponse()

		case packet.PktStoreResp:
			processingError = c.processStoreResponse()

		case packet.PktMkdirResp:
			processingError = c.processMkdirResponse()

		case packet.PktDeleteResp:
			processingError = c.processDeleteResponse()

		case packet.PktWriteResp:
			processingError = c.processWriteResponse()
		}

		if processingError != nil {
			c.logger.Error("net-process", processingError)
			c.fatalInternalError(err)
			return
		}
	}
}

func (c *RemoteSource) processWriteResponse() error {
	var writeResp packet.WriteResp
	err := c.transiever.GetWriteResp(&writeResp)
	if err != nil {
		return err
	}

	c.dispatchCallResponse(writeResp.ID, writeResp)
	return nil
}

func (c *RemoteSource) processDeleteResponse() error {
	var deleteResp packet.DeleteResp
	err := c.transiever.GetDeleteResp(&deleteResp)
	if err != nil {
		return err
	}

	c.dispatchCallResponse(deleteResp.ID, deleteResp)
	return nil
}

func (c *RemoteSource) processMkdirResponse() error {
	var mkdirResp packet.MkdirResp
	err := c.transiever.GetMkdirResp(&mkdirResp)
	if err != nil {
		return err
	}

	c.dispatchCallResponse(mkdirResp.ID, mkdirResp)
	return nil
}

func (c *RemoteSource) processStoreResponse() error {
	var storeResp packet.StoreResp
	err := c.transiever.GetStoreResp(&storeResp)
	if err != nil {
		return err
	}

	c.dispatchCallResponse(storeResp.ID, storeResp)
	return nil
}

func (c *RemoteSource) processFetchResponse() error {
	var fetchResponse packet.FetchResp
	err := c.transiever.GetFetchResp(&fetchResponse)
	if err != nil {
		return err
	}

	c.dispatchCallResponse(fetchResponse.ID, fetchResponse)
	return nil
}

func (c *RemoteSource) processListResponse() error {
	var listResponse packet.ListResp
	err := c.transiever.GetListResp(&listResponse)
	if err != nil {
		return err
	}

	c.dispatchCallResponse(listResponse.ID, listResponse)
	return nil
}

func (c *RemoteSource) processReadMetaResponse() error {
	var readMetaResponse packet.ReadMetaResp
	err := c.transiever.GetReadMetaResp(&readMetaResponse)
	if err != nil {
		return err
	}

	c.dispatchCallResponse(readMetaResponse.ID, readMetaResponse)
	return nil
}

func (c *RemoteSource) processLookupResponse() error {
	var lookupResponse packet.LookupResp
	err := c.transiever.GetLookupResp(&lookupResponse)
	if err != nil {
		return err
	}

	c.dispatchCallResponse(lookupResponse.ID, lookupResponse)
	return nil
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

// Latency returns the latency of the connection in nanoseconds.
func (c *RemoteSource) Latency() int64 {
	return c.latency.Nanoseconds()
}
