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
	"strconv"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/twitchyliquid64/nugget/inodeFactory"
	"github.com/twitchyliquid64/nugget/logger"
	"github.com/twitchyliquid64/nugget/nugg/client"
	"github.com/twitchyliquid64/nugget/nuggtofuse"
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

	l := logger.New(os.Stdout, os.Stderr)
	fatalErrChan := make(chan error)

	c, err := client.Open(connectAddrVar, certPemPathVar, keyPemPathVar, caCertPemPathVar, l, fatalErrChan)
	if err != nil {
		l.Error("main", "Could not connect to remote: ", err)
		os.Exit(1)
	}
	defer c.Close()

	fuseConn := doMount(flag.Arg(0), l, c, fatalErrChan)
	defer fuseConn.Close()

	waitInterrupt(fatalErrChan, l)
	fuse.Unmount(flag.Arg(0))
}

func doMount(mountpoint string, l *logger.Logger, provider *client.RemoteSource, fatalErrChan chan error) *fuse.Conn {
	inodeSource := inodeFactory.MakePathAwareFactory()

	mainFS := nuggtofuse.Make(provider, inodeSource, l)
	sysFS := sysstatfs.Make(inodeSource)
	sysFS.SetComputedVariable("ok", func() []byte { return []byte(boolToIntString(provider.Ready())) })
	sysFS.SetComputedVariable("latency", func() []byte { return []byte(strconv.FormatInt(provider.Latency(), 10)) })

	mainFS.SetOverride("sys", sysFS)

	//Create the mount
	c, err := mount(mountpoint)
	if err != nil {
		l.Error("mount", err)
		os.Exit(1)
	}

	go fsServeRoutine(c, fatalErrChan, mainFS)

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		l.Error("fs-serve", err)
		os.Exit(1)
	}
	return c
}

func fsServeRoutine(c *fuse.Conn, fatalError chan error, fsBackend fs.FS) {
	err := fs.Serve(c, fsBackend)
	if err != nil {
		fatalError <- err
	}
}

func mount(mountpoint string) (*fuse.Conn, error) {
	return fuse.Mount(
		mountpoint,
		fuse.FSName("nugg"),
		fuse.Subtype("nuggetfs"),
		fuse.LocalVolume(),
		fuse.VolumeName("nugg"),
	)
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

func boolToIntString(in bool) string {
	if in {
		return "1"
	}
	return "0"

}
