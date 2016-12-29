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
	"github.com/twitchyliquid64/nugget/nuggdb"
	"github.com/twitchyliquid64/nugget/nuggtofuse"
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
	flags()
	c, provider, fatalErrChan := doMount()
	defer c.Close()
	defer provider.Close()
	waitInterrupt(flag.Arg(0), fatalErrChan)
}

func doMount() (*fuse.Conn, nugget.DataSourceSink, chan error) {
	//Initialize the filesystem backend
	provider, err := nuggdb.Create(flag.Arg(1))
	if err != nil {
		log.Fatal("FS init failure: ", err)
	}
	mainFS := nuggtofuse.Make(provider, &inodeFactory.PathAwareFactory{LastIssuedInode: 1})

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

func waitInterrupt(mountPath string, fatalErrChan chan error) {
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
	fuse.Unmount(mountPath)
	// main()'s defer c.Close should do final cleanup
}
