package main

// nugg is a client for remote nuggFS.
// It needs to be given an address, a CA certificate, and a client certificate/key.
// It will connect to the remote using github.com/twitchyliquid64/nugget/nugg/client
// Which it will translate into a fuse filesystem using github.com/twitchyliquid64/nugget/nuggtofuse

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/twitchyliquid64/nugget/logger"
	"github.com/twitchyliquid64/nugget/nugg/client"
)

var connectAddrVar string
var caCertPemPathVar string
var certPemPathVar string
var keyPemPathVar string

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s <path-to-mountpoint>\n", os.Args[0])
	flag.PrintDefaults()
}

func flags() {
	flag.StringVar(&connectAddrVar, "addr", "localhost:27298", "Address of the remote nuggFS source to connect to, formatted <IP>:<port>")
	flag.StringVar(&caCertPemPathVar, "cacert", "ca.pem", "Path to the PEM-formatted authority certificate")
	flag.StringVar(&certPemPathVar, "cert", "cert.pem", "Path to the PEM-formatted client certificate")
	flag.StringVar(&keyPemPathVar, "key", "key.pem", "Path to the PEM-formatted client key")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	checkCertFiles()
}

func main() {
	flags()

	l := logger.New(os.Stdout, os.Stderr)
	fatalErrChan := make(chan error)

	c, err := client.Open(connectAddrVar, certPemPathVar, keyPemPathVar, caCertPemPathVar, l, fatalErrChan)
	if err != nil {
		l.Error("main", "Could not connect to remote: ", err)
		os.Exit(1)
	}

	n, err := c.Lookup("/")
	fmt.Println(n, err)

	meta, err := c.ReadMeta(n)
	fmt.Println(meta, err)

	waitInterrupt(fatalErrChan, l)
}

func checkCertFiles() {
	if !fileExists(caCertPemPathVar) {
		fmt.Fprintf(os.Stderr, "Err: Could not stat '%s'\n", caCertPemPathVar)
		os.Exit(1)
	}
	if !fileExists(certPemPathVar) {
		fmt.Fprintf(os.Stderr, "Err: Could not stat '%s'\n", certPemPathVar)
		os.Exit(1)
	}
	if !fileExists(keyPemPathVar) {
		fmt.Fprintf(os.Stderr, "Err: Could not stat '%s'\n", keyPemPathVar)
		os.Exit(1)
	}
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func waitInterrupt(fatalErrChan chan error, l *logger.Logger) {
	sig := make(chan os.Signal, 2)
	done := make(chan bool, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		done <- true
	}()

	select {
	case <-done:
		l.Info("main", "Recieved interrupt, shutting down.")
	case err := <-fatalErrChan:
		l.Error("main", "Fatal internal error: ", err)
	}
}
