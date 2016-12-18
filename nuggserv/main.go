package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/twitchyliquid64/nugget/nuggserv/serv"
)

var listenerAddrVar string
var caCertPemPathVar string
var certPemPathVar string
var keyPemPathVar string

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
}

func flags() {
	flag.StringVar(&listenerAddrVar, "listen", ":27298", "Network address to bind to, formatted <IP>:<port>")
	flag.StringVar(&caCertPemPathVar, "cacert", "ca.pem", "Path to the PEM-formatted authority certificate")
	flag.StringVar(&certPemPathVar, "cert", "cert.pem", "Path to the PEM-formatted server certificate")
	flag.StringVar(&keyPemPathVar, "key", "key.pem", "Path to the PEM-formatted server key")
	flag.Usage = usage
	flag.Parse()

	if listenerAddrVar == "" {
		usage()
		os.Exit(1)
	}
}

func main() {
	flags()
	checkCertFiles()
	_, err := serv.NewServer(listenerAddrVar, certPemPathVar, keyPemPathVar, caCertPemPathVar)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing network: %s\n", err)
		os.Exit(1)
	}

	time.Sleep(time.Second * 10)

	// Start our core internals: PathStore, NodeStore, ChunkStore
	// TODO: Implement these essentials

	// Start our routing internals : component mapping requests to functions with IDs
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
