package main

import (
	"flag"
	"fmt"
	"os"
)

var listenerAddrVar string

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s <path-to-mountpoint>\n", os.Args[0])
	flag.PrintDefaults()
}

func flags() {
	flag.StringVar(&listenerAddrVar, "listen", ":27298", "Network address to bind to, formatted <IP>:<port>")
	flag.Usage = usage
	flag.Parse()

	if listenerAddrVar == "" {
		usage()
		os.Exit(1)
	}
}

func main() {
	flags()

	// Start our core internals: PathStore, NodeStore, ChunkStore
	// TODO: Implement these essentials

	// Start our routing internals : component mapping requests to functions with IDs

	// Start our network stack - TLS init, connection manager, listener
}
