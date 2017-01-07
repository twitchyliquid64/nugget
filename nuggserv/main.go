package main

// nuggserv implements a nuggFS network server.

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/twitchyliquid64/nugget/logger"
	"github.com/twitchyliquid64/nugget/nuggdb"
	"github.com/twitchyliquid64/nugget/nuggserv/serv"
)

var listenerAddrVar string
var caCertPemPathVar string
var certPemPathVar string
var keyPemPathVar string

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s <path-to-data-dir>\n", os.Args[0])
	flag.PrintDefaults()
}

func flags() {
	flag.StringVar(&listenerAddrVar, "listen", ":27298", "Network address to bind to, formatted <IP>:<port>")
	flag.StringVar(&caCertPemPathVar, "cacert", "ca.pem", "Path to the PEM-formatted authority certificate")
	flag.StringVar(&certPemPathVar, "cert", "cert.pem", "Path to the PEM-formatted server certificate")
	flag.StringVar(&keyPemPathVar, "key", "key.pem", "Path to the PEM-formatted server key")
	flag.Usage = usage
	flag.Parse()

	if listenerAddrVar == "" || flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}
}

func main() {
	flags()
	checkCertFiles()
	l := logger.New(os.Stdout, os.Stderr)

	provider, err := nuggdb.Create(flag.Arg(0), l)
	if err != nil {
		l.Error("server", "Error initializing data storage: ", err)
		os.Exit(1)
	}

	s, err := serv.NewServer(listenerAddrVar, certPemPathVar, keyPemPathVar, caCertPemPathVar, provider, l)

	if err != nil {
		l.Error("server", "Error initializing network: ", err)
		os.Exit(1)
	} else {
		l.Info("server", "Started listening on ", listenerAddrVar)
	}
	defer s.Close()

	// Start our core internals: PathStore, NodeStore, ChunkStore
	// TODO: Implement these essentials

	// Start our routing internals : component mapping requests to functions with IDs

	fatalErrChan := make(chan error)
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
