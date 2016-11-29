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
	"github.com/twitchyliquid64/nugget/sysstatfs"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT\n", os.Args[0])
	flag.PrintDefaults()
}

func flags() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
}

func main() {
	flags()

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

	waitInterrupt(flag.Arg(0))
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
