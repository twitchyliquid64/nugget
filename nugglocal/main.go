package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/twitchyliquid64/nugget"
	"github.com/twitchyliquid64/nugget/inodeFactory"
	"github.com/twitchyliquid64/nugget/logger"
	"github.com/twitchyliquid64/nugget/nuggdb"
	"github.com/twitchyliquid64/nugget/nuggtofuse"
	"github.com/twitchyliquid64/nugget/sysstatfs"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s <path-to-mountpoint> <path-to-data-dir>\n", os.Args[0])
	flag.PrintDefaults()
}

func flags() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}
}

func main() {
	l := logger.New(os.Stdout, os.Stderr)
	flags()
	c, provider, fatalErrChan := doMount(l)
	defer provider.Close()
	waitInterrupt(fatalErrChan)

	//Close down
	fuse.Unmount(flag.Arg(0))
	c.Close()
}

func doMount(l *logger.Logger) (*fuse.Conn, nugget.DataSourceSink, chan error) {
	inodeSource := inodeFactory.MakePathAwareFactory()

	//Initialize the filesystem backend
	provider, err := nuggdb.Create(flag.Arg(1), l)
	if err != nil {
		log.Fatal("FS init failure: ", err)
	}
	mainFS := nuggtofuse.Make(provider, inodeSource, l)

	sysFS := sysstatfs.Make(inodeSource)
	mainFS.SetOverride("sys", sysFS)

	//Create the mount
	c, err := mount(flag.Arg(0))
	if err != nil {
		log.Fatal("Mount failure: ", err)
	}

	fatalErrorChan := make(chan error)
	go fsServeRoutine(c, fatalErrorChan, mainFS)

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
	return c, provider, fatalErrorChan
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

func waitInterrupt(fatalErrChan chan error) {
	sig := make(chan os.Signal, 2)
	done := make(chan bool, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		done <- true
	}()

	select {
	case <-done:
		log.Println("Recieved signal, terminating.")
	case err := <-fatalErrChan:
		log.Println("FUSE error: ", err)
	}
}
