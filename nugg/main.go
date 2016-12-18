package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/twitchyliquid64/nugget/nugg/client"
	"github.com/twitchyliquid64/nugget/sysstatfs"
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

	_, err := client.Open(connectAddrVar, certPemPathVar, keyPemPathVar, caCertPemPathVar)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to remote: %s\n", err)
	}

	time.Sleep(10 * time.Second)
	//doMount()
	//waitInterrupt(flag.Arg(0))
}

func doMount() {
	c, err := mount(flag.Arg(0))
	defer c.Close()
	if err != nil {
		log.Fatal(err)
	}

	sysfs := sysstatfs.Make()

	go func() {
		err := fs.Serve(c, sysfs)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}

func mount(name string) (*fuse.Conn, error) {
	return fuse.Mount(
		name,
		fuse.FSName("mirrorclient"),
		fuse.Subtype("nuggetfs"),
		fuse.LocalVolume(),
		fuse.VolumeName("Mirrorclient"),
	)
}

func waitInterrupt(mountPath string) {
	sig := make(chan os.Signal, 2)
	done := make(chan bool, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		done <- true
	}()
	<-done
	log.Println("Now shutting down.")
	fuse.Unmount(mountPath)
	// Main's defer c.Close should do final cleanup
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
